package integrationtest

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRedisRecoveryAfterPauseAndConnectionDrop(t *testing.T) {
	env := NewTestEnv(t, nil)

	addr, dbIndex, err := redisAddrAndDB(env.Config.RedisURL)
	require.NoError(t, err)

	controlConn, err := openRedisConn(addr, 500*time.Millisecond)
	if err != nil {
		t.Skipf("skipping redis recovery test because redis is unavailable: %v", err)
	}
	defer controlConn.Close()

	if dbIndex > 0 {
		_, err = controlConn.Command("SELECT", strconv.Itoa(dbIndex))
		require.NoError(t, err)
	}

	targetConn, err := openRedisConn(addr, 500*time.Millisecond)
	require.NoError(t, err)
	defer targetConn.Close()

	if dbIndex > 0 {
		_, err = targetConn.Command("SELECT", strconv.Itoa(dbIndex))
		require.NoError(t, err)
	}

	pong, err := targetConn.Command("PING")
	require.NoError(t, err)
	require.Equal(t, "PONG", strings.ToUpper(pong.String()))

	targetClientID, err := targetConn.Command("CLIENT", "ID")
	require.NoError(t, err)
	require.Greater(t, targetClientID.Integer(), int64(0))

	killResult, err := controlConn.Command("CLIENT", "KILL", "ID", strconv.FormatInt(targetClientID.Integer(), 10))
	require.NoError(t, err)
	require.GreaterOrEqual(t, killResult.Integer(), int64(1), "expected target redis client to be terminated")

	_, err = targetConn.Command("PING")
	require.Error(t, err, "expected terminated redis connection to fail before reconnect")

	reconnected, err := openRedisConn(addr, 500*time.Millisecond)
	require.NoError(t, err)
	defer reconnected.Close()

	if dbIndex > 0 {
		_, err = reconnected.Command("SELECT", strconv.Itoa(dbIndex))
		require.NoError(t, err)
	}

	pongAfterReconnect, err := reconnected.Command("PING")
	require.NoError(t, err)
	require.Equal(t, "PONG", strings.ToUpper(pongAfterReconnect.String()))

	// Simulate temporary outage behavior with Redis command pause.
	_, err = controlConn.Command("CLIENT", "PAUSE", "1200", "ALL")
	require.NoError(t, err)

	degradedConn, err := openRedisConn(addr, 250*time.Millisecond)
	require.NoError(t, err)
	defer degradedConn.Close()

	if dbIndex > 0 {
		_, _ = degradedConn.Command("SELECT", strconv.Itoa(dbIndex))
	}

	_, err = degradedConn.Command("PING")
	require.Error(t, err, "expected redis command failure/timeout during pause window")

	time.Sleep(1300 * time.Millisecond)

	pongAfterResume, err := controlConn.Command("PING")
	require.NoError(t, err)
	require.Equal(t, "PONG", strings.ToUpper(pongAfterResume.String()))
}

type redisConn struct {
	conn   net.Conn
	reader *bufio.Reader
	timeout time.Duration
}

func openRedisConn(addr string, timeout time.Duration) (*redisConn, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	return &redisConn{
		conn:   conn,
		reader: bufio.NewReader(conn),
		timeout: timeout,
	}, nil
}

func (c *redisConn) Close() error {
	return c.conn.Close()
}

func (c *redisConn) Command(args ...string) (redisValue, error) {
	if err := c.conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		return redisValue{}, err
	}
	if err := writeRESPArray(c.conn, args); err != nil {
		return redisValue{}, err
	}
	return readRESPValue(c.reader)
}

type redisValue struct {
	kind  byte
	str   string
	ival  int64
	items []redisValue
}

func (v redisValue) String() string {
	return v.str
}

func (v redisValue) Integer() int64 {
	return v.ival
}

func writeRESPArray(conn net.Conn, args []string) error {
	if _, err := fmt.Fprintf(conn, "*%d\r\n", len(args)); err != nil {
		return err
	}
	for _, arg := range args {
		if _, err := fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(arg), arg); err != nil {
			return err
		}
	}
	return nil
}

func readRESPValue(r *bufio.Reader) (redisValue, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return redisValue{}, err
	}

	line, err := readLineCRLF(r)
	if err != nil {
		return redisValue{}, err
	}

	switch prefix {
	case '+':
		return redisValue{kind: prefix, str: line}, nil
	case '-':
		return redisValue{}, fmt.Errorf("redis error: %s", line)
	case ':':
		n, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return redisValue{}, err
		}
		return redisValue{kind: prefix, ival: n}, nil
	case '$':
		length, err := strconv.Atoi(line)
		if err != nil {
			return redisValue{}, err
		}
		if length < 0 {
			return redisValue{kind: prefix}, nil
		}
		buf := make([]byte, length+2)
		if _, err := r.Read(buf); err != nil {
			return redisValue{}, err
		}
		return redisValue{kind: prefix, str: string(buf[:length])}, nil
	case '*':
		count, err := strconv.Atoi(line)
		if err != nil {
			return redisValue{}, err
		}
		items := make([]redisValue, 0, count)
		for i := 0; i < count; i++ {
			item, err := readRESPValue(r)
			if err != nil {
				return redisValue{}, err
			}
			items = append(items, item)
		}
		return redisValue{kind: prefix, items: items}, nil
	default:
		return redisValue{}, fmt.Errorf("unexpected redis RESP prefix: %q", prefix)
	}
}

func readLineCRLF(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), nil
}

func redisAddrAndDB(redisURL string) (string, int, error) {
	u, err := url.Parse(redisURL)
	if err != nil {
		return "", 0, err
	}

	host := u.Host
	if !strings.Contains(host, ":") {
		host += ":6379"
	}

	dbIndex := 0
	path := strings.TrimPrefix(u.Path, "/")
	if path != "" {
		n, err := strconv.Atoi(path)
		if err != nil {
			return "", 0, fmt.Errorf("invalid redis DB index in URL: %w", err)
		}
		dbIndex = n
	}

	return host, dbIndex, nil
}
