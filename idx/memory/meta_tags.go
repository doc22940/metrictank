package memory

import (
	"fmt"
	"hash"
	"hash/fnv"
	"sort"
	"strings"
)

// the collision avoidance window defines how many times we try to find a higher
// slot that's free if two record hashes collide
const collisionAvoidanceWindow = 3

// the function we use to get the hash for hashing the meta records
// it can be replaced for mocking in tests
var queryHash func() hash.Hash32

func init() {
	queryHash = fnv.New32a
}

type metaTagRecord struct {
	metaTags []kv
	queries  []expression
}

// list of meta records keyed by a unique identifier used as ID
type metaTagRecords map[uint32]metaTagRecord

// index structure keyed by tag -> value -> list of meta record IDs
type metaTagValue map[string][]uint32
type metaTagIndex map[string]metaTagValue

func (m metaTagIndex) deleteRecord(keyValue kv, recordId uint32) {
	if values, ok := m[keyValue.key]; ok {
		if recordIds, ok := values[keyValue.value]; ok {
			for i := 0; i < len(recordIds); i++ {
				if recordIds[i] == recordId {
					// no need to keep the order
					recordIds[i] = recordIds[len(recordIds)-1]
					values[keyValue.value] = recordIds[:len(recordIds)-1]

					// no id should ever be present more than once
					return
				}
			}
		}
	}
}

func (m metaTagIndex) insertRecord(keyValue kv, recordId uint32) {
	var values metaTagValue
	var ok bool

	if values, ok = m[keyValue.key]; !ok {
		values = make(metaTagValue)
		m[keyValue.key] = values
	}

	values[keyValue.value] = append(values[keyValue.value], recordId)
}

// metaTagRecordFromStrings takes two slices of strings, parses them and returns a metaTagRecord
// The first slice of strings has the meta tags & values
// The second slice has the tag query expressions which the meta tags & values refer to
// On parsing error the second returned value is an error, otherwise it is nil
func metaTagRecordFromStrings(metaTags []string, tagQueryExpressions []string) (metaTagRecord, error) {
	record := metaTagRecord{
		metaTags: make([]kv, 0, len(metaTags)),
		queries:  make([]expression, 0, len(tagQueryExpressions)),
	}

	if len(tagQueryExpressions) == 0 {
		return record, fmt.Errorf("Requiring at least one tag query expression, 0 given")
	}

	for _, tag := range metaTags {
		tagSplits := strings.SplitN(tag, "=", 2)
		if len(tagSplits) < 2 {
			return record, fmt.Errorf("Missing \"=\" sign in tag %s", tag)
		}
		if len(tagSplits[0]) == 0 || len(tagSplits[1]) == 0 {
			return record, fmt.Errorf("Tag/Value cannot be empty in %s", tag)
		}

		record.metaTags = append(record.metaTags, kv{key: tagSplits[0], value: tagSplits[1]})
	}

	for _, query := range tagQueryExpressions {
		parsed, err := parseExpression(query)
		if err != nil {
			return record, err
		}
		record.queries = append(record.queries, parsed)
	}

	return record, nil
}

func (m *metaTagRecord) metaTagStrings(builder *strings.Builder) []string {
	res := make([]string, len(m.metaTags))

	for i, tag := range m.metaTags {
		tag.stringIntoBuilder(builder)
		res[i] = builder.String()
		builder.Reset()
	}

	return res
}

func (m *metaTagRecord) queryStrings(builder *strings.Builder) []string {
	res := make([]string, len(m.queries))

	for i, query := range m.queries {
		query.stringIntoBuilder(builder)
		res[i] = builder.String()
		builder.Reset()
	}

	return res
}

// hashQueries generates a hash of all the queries in the record
func (m *metaTagRecord) hashQueries() uint32 {
	builder := strings.Builder{}
	for _, query := range m.queries {
		query.stringIntoBuilder(&builder)

		// trailing ";" doesn't matter, this is only hash input
		builder.WriteString(";")
	}

	h := queryHash()
	h.Write([]byte(builder.String()))
	return h.Sum32()
}

// sortQueries sorts all the queries first by key, then by value, then by
// operator. The order doesn't matter, it only needs to be consistent
func (m *metaTagRecord) sortQueries() {
	sort.Slice(m.queries, func(i, j int) bool {
		if m.queries[i].key == m.queries[j].key {
			if m.queries[i].value == m.queries[j].value {
				return m.queries[i].operator < m.queries[j].operator
			}
			return m.queries[i].value < m.queries[j].value
		}
		return m.queries[i].key < m.queries[j].key
	})
}

// matchesQueries compares another tag record's queries to this
// one's queries. Returns true if they are equal, otherwise false.
// It is assumed that all the queries are already sorted
func (m *metaTagRecord) matchesQueries(other metaTagRecord) bool {
	if len(m.queries) != len(other.queries) {
		return false
	}

	for i, query := range m.queries {
		if query.key != other.queries[i].key {
			return false
		}

		if query.operator != other.queries[i].operator {
			return false
		}

		if query.value != other.queries[i].value {
			return false
		}
	}

	return true
}

// hasMetaTags returns true if the meta tag record has one or more
// meta tags, otherwise it returns false
func (m *metaTagRecord) hasMetaTags() bool {
	return len(m.metaTags) > 0
}

// upsert inserts or updates a meta tag record according to the given specifications
// it uses the set of tag query expressions as the identity of the record, if a record with the
// same identity is already present then its meta tags get updated to the specified ones.
// If the new record contains no meta tags, then the update is equivalent to a delete.
// Those are the return values:
// 1) The id at which the new record got inserted
// 2) Pointer to the inserted metaTagRecord
// 3) The id of the record that has been replaced if an update was performed
// 4) Pointer to the metaTagRecord that has been replaced if an update was performed, otherwise nil
// 5) Error if an error occurred, otherwise it's nil
func (m metaTagRecords) upsert(metaTags []string, tagQueryExpressions []string) (uint32, *metaTagRecord, uint32, *metaTagRecord, error) {
	record, err := metaTagRecordFromStrings(metaTags, tagQueryExpressions)
	if err != nil {
		return 0, nil, 0, nil, err
	}

	record.sortQueries()
	id := record.hashQueries()
	var oldRecord *metaTagRecord
	var oldId uint32

	// loop over existing records, starting from id, trying to find one that has
	// the exact same queries as the one we're upserting
	for i := uint32(0); i < collisionAvoidanceWindow; i++ {
		if existingRecord, ok := m[id+i]; ok {
			if record.matchesQueries(existingRecord) {
				oldRecord = &existingRecord
				oldId = id + i
				delete(m, oldId)
				break
			}
		}
	}

	if !record.hasMetaTags() {
		return 0, &record, oldId, oldRecord, nil
	}

	// now find the best position to insert the new/updated record, starting from id
	for i := uint32(0); i < collisionAvoidanceWindow; i++ {
		// if we find a free slot, then insert the new record there
		if _, ok := m[id]; !ok {
			m[id] = record

			return id, &record, oldId, oldRecord, nil
		}

		id++
	}

	return 0, nil, 0, nil, fmt.Errorf("MetaTagRecordUpsert: Unable to find a slot to insert record")
}
