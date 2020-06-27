## Main feature

Provide a solution to do horizontal scaling of PostgreSQL DBs by sharding the data between distinct PostgreSQL instances.

## Initial design constraints

- fixed number of shards: i.e. it is not possible to add/remove shards after initial configuration
- user must submit a virtual schema so that Matriarch, when receiving a SQL command, knows which shards to target to execute the operation
- user should never use sequences as column identifiers (serial) but should use UUIDs instead, as sequences do not fit well with a distributed DB

## Concepts:

- Keyspace = logical database that maps to multiple PGSQL databases, each one owned by a different shard. A keyspace appears as a single database from the standpoint of the application
- Each shard is a PGSQL cluster, composed of a primary and secondaries, owning a portion of the keyspace (really a range of keyspace ID values in the keyspace). Each shard contains sharded tables (content scattered amongst shards) and reference tables (same data copied everywhere, read only tables)

  - Shard naming: Example: shard names have the following characteristics:

    - They represent a range, where the left number is included, but the right is not.
    - Their notation is hexadecimal.
    - They are left justified.
    - prefix means: anything less than the right value.
    - postfix means: anything greater than or equal to the LHS value.
    - A plain - denotes the full keyrange

  An example of shard naming configuration is `Customer/-80`, meaning this shard will own all rows whose keyspaceID calculated using the Primary Vindex function is lower than x80000000000000000000 (SHA-1 produces 20 bytes long hashes).

- The keyspace ID is the value that is used to decide on which shard a given row lives. Range-based Sharding refers to creating shards that each cover a particular range of keyspace ID (for all the tables inside the shard database). The keyspace ID itself is computed using a function of some column in your data, such as the user ID. Matriarch uses a hash function (vindex) to perform this mapping. The keyspace ID is a concept that is internal to Matriarch. The application does not need to know anything about it. There is no physical column that stores the actual keyspace ID. This value is computed as needed.
- Virtual Index: A Vindex maps column values to keyspace IDs. A Vindex provides a way to map a column value to a keyspace ID. This mapping can be used to identify the location of a row. A table can have multiple Vindexes.
  - The Primary Vindex: it is analogous to a database primary key. Every sharded table must have one defined. A Primary Vindex must be unique: given an input value, it must produce a single keyspace ID. This unique mapping will be used at the time of insert to decide the target shard for a row. Conceptually, this is also equivalent to the NoSQL Sharding Key, and we often refer to the Primary Vindex as the Sharding Key. Uniqueness for a Primary Vindex does not mean that the column has to be a primary key or unique in the PostgreSQL schema. You can have multiple rows that map to the same keyspace ID. The Vindex uniqueness constraint is only used to make sure that all rows for a keyspace ID live in the same shard. Primary Vindex in Vitess not only defines the Sharding Key, it also decides the Sharding Scheme.
  - Secondary Vindexes are additional vindexes you can define against other columns of a table offering you optimizations for WHERE clauses that do not use the Primary Vindex. Secondary Vindexes return a single or a limited set of keyspace IDs which will allow Matriarch to only target shards where the relevant data is present. In the absence of a Secondary Vindex, Matriarch would have to send the query to all shards. Secondary Vindexes are also commonly known as cross-shard indexes. It is important to note that Secondary Vindexes are only for making routing decisions. The underlying database shards will most likely need traditional indexes on those same columns.
  - A Unique Vindex is one that yields at most one keyspace ID for a given input. Knowing that a Vindex is Unique is useful because VTGate can push down some complex queries into VTTablet if it knows that the scope of that query cannot exceed a shard. Uniqueness is also a prerequisite for a Vindex to be used as Primary Vindex.
  - A NonUnique Vindex is analogous to a database non-unique index. It is a secondary index for searching by an alternate WHERE clause. An input value could yield multiple keyspace IDs, and rows could be matched from multiple shards. For example, if a table has a name column that allows duplicates, you can define a cross-shard NonUnique Vindex for it, and this will let you efficiently search for users that match a certain name.

## How it works

Matriarch

- Reads the configuration file on startup, validates it and then builds an in memory representation of the available pgsql backends
- Reads the vschema file, validates it and then builds an in memory representation of the shards using information about pgsql backends found in the conf file
- Establishes a connection to each of the shards
  - On initial connection, checks if DB exists on each shard, otherwise it creates the DB
  - Connection lifecycle mgmt: if a shard dies and one of its replicas takes over, we need to reconnect

- On INSERT, Matriarch generate the keyspaceID for the new row, by SHA-1 the string resulting of the concatenation of values, separated by of the columns composing Primary Vindex of the table, and then finds the Shard owning the portion of the KeyspaceID in which the just calculated KeyspaceID falls inside. Updates to columns composing the Primary Vindex are allowed if the update doesn't result in a change of shard
- how to make sure that queries using secondary indexes produce keyspaceIDs that fall within the range of the shard really storing the desired data?
  - Secondary indexes are expressed on columns, and the system maintains (in-memory or elsewhere) structures that map combination of those columns values and shards that hold those rows. E.g. secondary index on `customer.location, customer.age`, then the structure maps existing values to shards holding rows with those values: map where key is the `WHERE` clause `customer.location="IT" && customer.age=25` and value is the array of shards owning rows which have those values for those columns.
  - On INSERT/UPDATE/DELETE, secondary indexes must be updated, meaning that this map must be updated as well.

## Future features

- Can the schema be read directly from PGSQL by crafting a PGSQL extension that allows devs to add metadata to the schema without having to manage a separate schema into another tool?
- Dynamic sharding
  - add a shard and split the original keyspace range in two, so that the new shard can take half and replicate half the data
- Connection pooling
  - https://scalegrid.io/blog/postgresql-connection-pooling-part-1-pros-and-cons/
  - https://scalegrid.io/blog/postgresql-connection-pooling-part-2-pgbouncer/
  - https://scalegrid.io/blog/postgresql-connection-pooling-part-3-pgpool-ii/
- Instrumentation of queries with metrics
  - parse incoming statements and extract basic metrics such as number of calls?

Tests and benchmarking

- pgbench
