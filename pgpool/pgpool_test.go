// Code taken and adapted from https://github.com/jackc/pgx/tree/master/pgxpool
package pgpool

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConnString() string {
	var connString string
	if connString = os.Getenv("PGX_TEST_DATABASE"); strings.TrimSpace(connString) == "" {
		connString = "postgres://localhost:5432/matriarch_test"
	}
	return connString
}

func TestConnect(t *testing.T) {
	t.Parallel()
	connString := testConnString()
	pool, err := Connect(context.Background(), connString)
	require.NoError(t, err)
	pool.Close()
}

func TestConnectConfig(t *testing.T) {
	t.Parallel()
	connString := testConnString()
	config, err := ParseConfig(connString)
	require.NoError(t, err)
	pool, err := ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	assertConfigsEqual(t, config, pool.Config(), "Pool.Config() returns original config")
	pool.Close()
}

func TestParseConfigExtractsPoolArguments(t *testing.T) {
	t.Parallel()

	config, err := ParseConfig("pool_max_conns=42 pool_min_conns=1")
	assert.NoError(t, err)
	assert.EqualValues(t, 42, config.MaxConns)
	assert.EqualValues(t, 1, config.MinConns)
	assert.NotContains(t, config.ConnConfig.RuntimeParams, "pool_max_conns")
	assert.NotContains(t, config.ConnConfig.RuntimeParams, "pool_min_conns")
}

func TestConnectCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pool, err := Connect(ctx, testConnString())
	assert.Nil(t, pool)
	assert.Equal(t, context.Canceled, err)
}

func TestLazyConnect(t *testing.T) {
	t.Parallel()

	config, err := ParseConfig(testConnString())
	assert.NoError(t, err)
	config.LazyConnect = true

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pool, err := ConnectConfig(ctx, config)
	assert.NoError(t, err)

	_, err = pool.Exec(ctx, "SELECT 1")
	assert.Equal(t, context.Canceled, err)
}

func TestConnectConfigRequiresConnConfigFromParseConfig(t *testing.T) {
	t.Parallel()

	config := &Config{}

	require.PanicsWithValue(t, "config must be created by ParseConfig", func() { ConnectConfig(context.Background(), config) })
}

func TestConfigCopyReturnsEqualConfig(t *testing.T) {
	connString := "postgres://jack:secret@localhost:5432/mydb?application_name=pgxtest&search_path=myschema&connect_timeout=5"
	original, err := ParseConfig(connString)
	require.NoError(t, err)

	copied := original.Copy()

	assertConfigsEqual(t, original, copied, t.Name())
}

func TestConfigCopyCanBeUsedToConnect(t *testing.T) {
	connString := testConnString()
	original, err := ParseConfig(connString)
	require.NoError(t, err)

	copied := original.Copy()
	assert.NotPanics(t, func() {
		_, err = ConnectConfig(context.Background(), copied)
	})
	assert.NoError(t, err)
}

func TestPoolAcquireAndConnRelease(t *testing.T) {
	t.Parallel()

	pool, err := Connect(context.Background(), testConnString())
	require.NoError(t, err)
	defer pool.Close()

	c, err := pool.acquire(context.Background())
	require.NoError(t, err)
	c.Release()
}

func TestPoolAfterConnect(t *testing.T) {
	t.Parallel()

	config, err := ParseConfig(testConnString())
	require.NoError(t, err)

	config.AfterConnect = func(ctx context.Context, c *pgconn.PgConn) error {
		_ = c.Exec(ctx, "select 1")
		return nil
	}

	db, err := ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	defer db.Close()

	// var n int32
	// err = db.QueryRow(context.Background(), "ps1").Scan(&n)
	// require.NoError(t, err)
	// assert.EqualValues(t, 1, n)
}

func TestPoolBeforeAcquire(t *testing.T) {
	t.Parallel()

	config, err := ParseConfig(testConnString())
	require.NoError(t, err)

	acquireAttempts := 0

	config.BeforeAcquire = func(ctx context.Context, c *pgconn.PgConn) bool {
		acquireAttempts++
		return acquireAttempts%2 == 0
	}

	db, err := ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	defer db.Close()

	conns := make([]*connResource, 4)
	for i := range conns {
		conns[i], err = db.acquire(context.Background())
		assert.NoError(t, err)
	}

	for _, c := range conns {
		c.Release()
	}
	waitForReleaseToComplete()

	assert.EqualValues(t, 8, acquireAttempts)

	conns = db.acquireAllIdle(context.Background())
	assert.Len(t, conns, 2)

	for _, c := range conns {
		c.Release()
	}
	waitForReleaseToComplete()

	assert.EqualValues(t, 12, acquireAttempts)
}

func TestPoolAfterRelease(t *testing.T) {
	t.Parallel()

	config, err := ParseConfig(testConnString())
	require.NoError(t, err)

	afterReleaseCount := 0

	config.AfterRelease = func(c *pgconn.PgConn) bool {
		afterReleaseCount++
		return afterReleaseCount%2 == 1
	}

	db, err := ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	defer db.Close()

	connPIDs := map[uint32]struct{}{}

	for i := 0; i < 10; i++ {
		conn, err := db.acquire(context.Background())
		assert.NoError(t, err)
		connPIDs[conn.conn.PID()] = struct{}{}
		conn.Release()
		waitForReleaseToComplete()
	}

	assert.EqualValues(t, 5, len(connPIDs))
}

func TestPoolAcquireAllIdle(t *testing.T) {
	t.Parallel()

	db, err := Connect(context.Background(), testConnString())
	require.NoError(t, err)
	defer db.Close()

	conns := db.acquireAllIdle(context.Background())
	assert.Len(t, conns, 1)

	for _, c := range conns {
		c.Release()
	}
	waitForReleaseToComplete()

	conns = make([]*connResource, 3)
	for i := range conns {
		conns[i], err = db.acquire(context.Background())
		assert.NoError(t, err)
	}

	for _, c := range conns {
		if c != nil {
			c.Release()
		}
	}
	waitForReleaseToComplete()

	conns = db.acquireAllIdle(context.Background())
	assert.Len(t, conns, 3)

	for _, c := range conns {
		c.Release()
	}
}

func TestConnReleaseChecksMaxConnLifetime(t *testing.T) {
	t.Parallel()

	config, err := ParseConfig(testConnString())
	require.NoError(t, err)

	config.MaxConnLifetime = 250 * time.Millisecond

	db, err := ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	defer db.Close()

	c, err := db.acquire(context.Background())
	require.NoError(t, err)

	time.Sleep(config.MaxConnLifetime)

	c.Release()
	waitForReleaseToComplete()

	stats := db.Stat()
	assert.EqualValues(t, 0, stats.TotalConns())
}

func TestConnReleaseClosesBusyConn(t *testing.T) {
	t.Parallel()

	db, err := Connect(context.Background(), testConnString())
	require.NoError(t, err)
	defer db.Close()

	c, err := db.acquire(context.Background())
	require.NoError(t, err)

	res := c.conn.Exec(context.Background(), "select generate_series(1,10)")
	require.NotNil(t, res)

	c.Release()
	waitForReleaseToComplete()

	stats := db.Stat()
	assert.EqualValues(t, 0, stats.TotalConns())
}

func TestPoolBackgroundChecksMaxConnLifetime(t *testing.T) {
	t.Parallel()

	config, err := ParseConfig(testConnString())
	require.NoError(t, err)

	config.MinConns = 0
	config.MaxConnLifetime = 100 * time.Millisecond
	config.HealthCheckPeriod = 100 * time.Millisecond

	db, err := ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	defer db.Close()

	c, err := db.acquire(context.Background())
	require.NoError(t, err)
	c.Release()
	time.Sleep(config.MaxConnLifetime + 100*time.Millisecond)

	stats := db.Stat()
	assert.EqualValues(t, 0, stats.TotalConns())
}

func TestPoolBackgroundChecksMaxConnIdleTime(t *testing.T) {
	t.Parallel()

	config, err := ParseConfig(testConnString())
	require.NoError(t, err)

	config.MinConns = 0
	config.MaxConnLifetime = 1 * time.Minute
	config.MaxConnIdleTime = 100 * time.Millisecond
	config.HealthCheckPeriod = 150 * time.Millisecond

	db, err := ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	defer db.Close()

	c, err := db.acquire(context.Background())
	require.NoError(t, err)
	c.Release()
	time.Sleep(config.HealthCheckPeriod + 50*time.Millisecond)

	stats := db.Stat()
	assert.EqualValues(t, 0, stats.TotalConns())
}

func TestPoolBackgroundChecksMinConns(t *testing.T) {
	config, err := ParseConfig(testConnString())
	require.NoError(t, err)

	config.HealthCheckPeriod = 100 * time.Millisecond
	config.MinConns = 2

	db, err := ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	defer db.Close()

	time.Sleep(config.HealthCheckPeriod + 100*time.Millisecond)

	stats := db.Stat()
	assert.EqualValues(t, 2, stats.TotalConns())
}

func TestPoolExec(t *testing.T) {
	t.Parallel()

	pool, err := Connect(context.Background(), testConnString())
	require.NoError(t, err)
	defer pool.Close()

	testExec(t, pool)
}

func TestConnReleaseDestroysClosedConn(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := Connect(ctx, testConnString())
	require.NoError(t, err)
	defer pool.Close()

	c, err := pool.acquire(ctx)
	require.NoError(t, err)

	err = c.conn.Close(ctx)
	require.NoError(t, err)

	assert.EqualValues(t, 1, pool.Stat().TotalConns())

	c.Release()
	waitForReleaseToComplete()

	assert.EqualValues(t, 0, pool.Stat().TotalConns())
}

// Conn.Release is an asynchronous process that returns immediately. There is no signal when the actual work is
// completed. To test something that relies on the actual work for Conn.Release being completed we must simply wait.
// This function wraps the sleep so there is more meaning for the callers.
func waitForReleaseToComplete() {
	time.Sleep(5 * time.Millisecond)
}

type execer interface {
	Exec(ctx context.Context, sql string) (*pgconn.MultiResultReader, error)
}

func testExec(t *testing.T, db execer) {
	results, err := db.Exec(context.Background(), "set time zone 'Europe/Paris'")
	require.NoError(t, err)
	require.NotNil(t, results)
}

func assertConfigsEqual(t *testing.T, expected, actual *Config, testName string) {
	if !assert.NotNil(t, expected) {
		return
	}
	if !assert.NotNil(t, actual) {
		return
	}

	// Can't test function equality, so just test that they are set or not.
	assert.Equalf(t, expected.AfterConnect == nil, actual.AfterConnect == nil, "%s - AfterConnect", testName)
	assert.Equalf(t, expected.BeforeAcquire == nil, actual.BeforeAcquire == nil, "%s - BeforeAcquire", testName)
	assert.Equalf(t, expected.AfterRelease == nil, actual.AfterRelease == nil, "%s - AfterRelease", testName)

	assert.Equalf(t, expected.MaxConnLifetime, actual.MaxConnLifetime, "%s - MaxConnLifetime", testName)
	assert.Equalf(t, expected.MaxConnIdleTime, actual.MaxConnIdleTime, "%s - MaxConnIdleTime", testName)
	assert.Equalf(t, expected.MaxConns, actual.MaxConns, "%s - MaxConns", testName)
	assert.Equalf(t, expected.MinConns, actual.MinConns, "%s - MinConns", testName)
	assert.Equalf(t, expected.HealthCheckPeriod, actual.HealthCheckPeriod, "%s - HealthCheckPeriod", testName)
	assert.Equalf(t, expected.LazyConnect, actual.LazyConnect, "%s - LazyConnect", testName)

	assertConnConfigsEqual(t, expected.ConnConfig, actual.ConnConfig, testName)
}

func assertConnConfigsEqual(t *testing.T, expected, actual *pgconn.Config, testName string) {
	if !assert.NotNil(t, expected) {
		return
	}
	if !assert.NotNil(t, actual) {
		return
	}

	// Can't test function equality, so just test that they are set or not.
	assert.Equalf(t, expected.Host, actual.Host, "%s - Host", testName)
	assert.Equalf(t, expected.Database, actual.Database, "%s - Database", testName)
	assert.Equalf(t, expected.Port, actual.Port, "%s - Port", testName)
	assert.Equalf(t, expected.User, actual.User, "%s - User", testName)
	assert.Equalf(t, expected.Password, actual.Password, "%s - Password", testName)
	assert.Equalf(t, expected.ConnectTimeout, actual.ConnectTimeout, "%s - ConnectTimeout", testName)
	assert.Equalf(t, expected.RuntimeParams, actual.RuntimeParams, "%s - RuntimeParams", testName)

	// Can't test function equality, so just test that they are set or not.
	assert.Equalf(t, expected.ValidateConnect == nil, actual.ValidateConnect == nil, "%s - ValidateConnect", testName)
	assert.Equalf(t, expected.AfterConnect == nil, actual.AfterConnect == nil, "%s - AfterConnect", testName)

	if assert.Equalf(t, expected.TLSConfig == nil, actual.TLSConfig == nil, "%s - TLSConfig", testName) {
		if expected.TLSConfig != nil {
			assert.Equalf(t, expected.TLSConfig.InsecureSkipVerify, actual.TLSConfig.InsecureSkipVerify, "%s - TLSConfig InsecureSkipVerify", testName)
			assert.Equalf(t, expected.TLSConfig.ServerName, actual.TLSConfig.ServerName, "%s - TLSConfig ServerName", testName)
		}
	}

	if assert.Equalf(t, len(expected.Fallbacks), len(actual.Fallbacks), "%s - Fallbacks", testName) {
		for i := range expected.Fallbacks {
			assert.Equalf(t, expected.Fallbacks[i].Host, actual.Fallbacks[i].Host, "%s - Fallback %d - Host", testName, i)
			assert.Equalf(t, expected.Fallbacks[i].Port, actual.Fallbacks[i].Port, "%s - Fallback %d - Port", testName, i)

			if assert.Equalf(t, expected.Fallbacks[i].TLSConfig == nil, actual.Fallbacks[i].TLSConfig == nil, "%s - Fallback %d - TLSConfig", testName, i) {
				if expected.Fallbacks[i].TLSConfig != nil {
					assert.Equalf(t, expected.Fallbacks[i].TLSConfig.InsecureSkipVerify, actual.Fallbacks[i].TLSConfig.InsecureSkipVerify, "%s - Fallback %d - TLSConfig InsecureSkipVerify", testName)
					assert.Equalf(t, expected.Fallbacks[i].TLSConfig.ServerName, actual.Fallbacks[i].TLSConfig.ServerName, "%s - Fallback %d - TLSConfig ServerName", testName)
				}
			}
		}
	}
}
