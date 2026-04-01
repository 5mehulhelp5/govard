package remote

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"govard/internal/engine"
)

type DotenvDBInfo struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type DotenvEnvironment struct {
	DB DotenvDBInfo
}

type dotenvDBProbePayload struct {
	DatabaseURL      string `json:"database_url"`
	DatabaseHost     string `json:"database_host"`
	DatabasePort     string `json:"database_port"`
	DatabaseUser     string `json:"database_user"`
	DatabasePassword string `json:"database_password"`
	DatabaseName     string `json:"database_name"`
	DBHost           string `json:"db_host"`
	DBPort           string `json:"db_port"`
	DBName           string `json:"db_name"`
	DBDatabase       string `json:"db_database"`
	DBUser           string `json:"db_user"`
	DBUsername       string `json:"db_username"`
	DBPassword       string `json:"db_password"`
	MysqlUser        string `json:"mysql_user"`
	MysqlDatabase    string `json:"mysql_database"`
	MysqlPassword    string `json:"mysql_password"`
	MysqlHost        string `json:"mysql_host"`
	MysqlPort        string `json:"mysql_port"`
}

func ProbeDotenvEnvironment(remoteName string, remoteCfg engine.RemoteConfig) (DotenvEnvironment, error) {
	remoteCommand := buildMagentoRemoteCommand(remoteCfg.Path, `php -r `+engine.ShellQuote(dotenvDBProbePHP))
	encoded, err := runRemoteCapture(remoteName, remoteCfg, remoteCommand)
	if err != nil {
		return DotenvEnvironment{}, err
	}
	return decodeDotenvEnvironmentPayload(encoded)
}

func decodeDotenvEnvironmentPayload(encoded string) (DotenvEnvironment, error) {
	trimmed := strings.TrimSpace(encoded)
	if trimmed == "" {
		return DotenvEnvironment{}, fmt.Errorf("remote probe returned empty payload")
	}

	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return DotenvEnvironment{}, fmt.Errorf("decode remote probe payload: %w", err)
	}

	var payload dotenvDBProbePayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return DotenvEnvironment{}, fmt.Errorf("parse remote probe payload: %w", err)
	}

	info, err := resolveDotenvDBInfo(payload)
	if err != nil {
		return DotenvEnvironment{}, err
	}

	return DotenvEnvironment{DB: info}, nil
}

func resolveDotenvDBInfo(payload dotenvDBProbePayload) (DotenvDBInfo, error) {
	if databaseURL := strings.TrimSpace(payload.DatabaseURL); databaseURL != "" {
		return parseDotenvDatabaseURL(databaseURL)
	}

	host := firstNonEmpty(payload.DatabaseHost, payload.DBHost, payload.MysqlHost)
	if host == "" {
		host = "127.0.0.1"
	}

	port := 3306
	portRaw := strings.TrimSpace(firstNonEmpty(payload.DatabasePort, payload.DBPort, payload.MysqlPort))
	if portRaw != "" {
		parsed, err := strconv.Atoi(portRaw)
		if err != nil || parsed <= 0 {
			return DotenvDBInfo{}, fmt.Errorf("invalid database port %q from remote dotenv", portRaw)
		}
		port = parsed
	}

	username := strings.TrimSpace(firstNonEmpty(payload.DatabaseUser, payload.DBUsername, payload.DBUser, payload.MysqlUser))
	database := strings.TrimSpace(firstNonEmpty(payload.DatabaseName, payload.DBDatabase, payload.DBName, payload.MysqlDatabase))
	if username == "" || database == "" {
		return DotenvDBInfo{}, fmt.Errorf("remote dotenv is missing database username or database name")
	}

	return DotenvDBInfo{
		Host:     host,
		Port:     port,
		Username: username,
		Password: firstNonEmpty(payload.DatabasePassword, payload.DBPassword, payload.MysqlPassword),
		Database: database,
	}, nil
}

func parseDotenvDatabaseURL(raw string) (DotenvDBInfo, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.Trim(trimmed, `"'`)
	if trimmed == "" {
		return DotenvDBInfo{}, fmt.Errorf("remote dotenv DATABASE_URL is empty")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return DotenvDBInfo{}, fmt.Errorf("parse remote DATABASE_URL: %w", err)
	}

	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	switch scheme {
	case "mysql", "mysql2", "mariadb":
	default:
		return DotenvDBInfo{}, fmt.Errorf("remote dotenv DATABASE_URL uses unsupported scheme %q", parsed.Scheme)
	}

	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		host = "127.0.0.1"
	}

	port := 3306
	if rawPort := strings.TrimSpace(parsed.Port()); rawPort != "" {
		parsedPort, parseErr := strconv.Atoi(rawPort)
		if parseErr != nil || parsedPort <= 0 {
			return DotenvDBInfo{}, fmt.Errorf("remote dotenv DATABASE_URL has invalid port %q", rawPort)
		}
		port = parsedPort
	}

	username := strings.TrimSpace(parsed.User.Username())
	password, _ := parsed.User.Password()
	database := strings.TrimSpace(strings.TrimPrefix(parsed.Path, "/"))
	if unescaped, unescapeErr := url.PathUnescape(database); unescapeErr == nil {
		database = unescaped
	}

	if username == "" || database == "" {
		return DotenvDBInfo{}, fmt.Errorf("remote dotenv DATABASE_URL is missing username or database name")
	}

	return DotenvDBInfo{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func ParseDotenvDatabaseURLForTest(raw string) (DotenvDBInfo, error) {
	return parseDotenvDatabaseURL(raw)
}

func ResolveDotenvDBInfoForTest(
	databaseURL string,
	databaseHost string,
	databasePort string,
	databaseUser string,
	databasePassword string,
	databaseName string,
	dbHost string,
	dbPort string,
	dbName string,
	dbDatabase string,
	dbUser string,
	dbUsername string,
	dbPassword string,
) (DotenvDBInfo, error) {
	return resolveDotenvDBInfo(dotenvDBProbePayload{
		DatabaseURL:      databaseURL,
		DatabaseHost:     databaseHost,
		DatabasePort:     databasePort,
		DatabaseUser:     databaseUser,
		DatabasePassword: databasePassword,
		DatabaseName:     databaseName,
		DBHost:           dbHost,
		DBPort:           dbPort,
		DBName:           dbName,
		DBDatabase:       dbDatabase,
		DBUser:           dbUser,
		DBUsername:       dbUsername,
		DBPassword:       dbPassword,
	})
}

const dotenvDBProbePHP = `
function govard_parse_env($path, &$vars) {
	if (!is_file($path)) {
		return;
	}
	$lines = @file($path, FILE_IGNORE_NEW_LINES);
	if ($lines === false) {
		return;
	}
	foreach ($lines as $line) {
		$line = trim((string)$line);
		if ($line === '' || $line[0] === '#') {
			continue;
		}
		if (isset($line[6]) && substr($line, 0, 7) === 'export ') {
			$line = substr($line, 7);
		}
		$pos = strpos($line, '=');
		if ($pos === false) {
			continue;
		}
		$key = trim(substr($line, 0, $pos));
		if ($key === '') {
			continue;
		}
		$value = trim(substr($line, $pos + 1));
		if ($value !== '') {
			$first = $value[0];
			$last = substr($value, -1);
			if (($first === '"' || $first === "'") && $last === $first && strlen($value) >= 2) {
				$value = substr($value, 1, -1);
				if ($first === '"') {
					$value = stripcslashes($value);
				}
			}
		}
		$vars[$key] = $value;
	}
}

$vars = [];
govard_parse_env('.env', $vars);
$appEnv = trim((string)($vars['APP_ENV'] ?? 'dev'));
if ($appEnv === '') {
	$appEnv = 'dev';
}
govard_parse_env('.env.local', $vars);
govard_parse_env('.env.' . $appEnv, $vars);
govard_parse_env('.env.' . $appEnv . '.local', $vars);

$out = [
	'database_url' => (string)($vars['DATABASE_URL'] ?? ''),
	'database_host' => (string)($vars['DATABASE_HOST'] ?? ''),
	'database_port' => (string)($vars['DATABASE_PORT'] ?? ''),
	'database_user' => (string)($vars['DATABASE_USER'] ?? ''),
	'database_password' => (string)($vars['DATABASE_PASSWORD'] ?? ''),
	'database_name' => (string)($vars['DATABASE_NAME'] ?? ''),
	'db_host' => (string)($vars['DB_HOST'] ?? ''),
	'db_port' => (string)($vars['DB_PORT'] ?? ''),
	'db_name' => (string)($vars['DB_NAME'] ?? ''),
	'db_database' => (string)($vars['DB_DATABASE'] ?? ''),
	'db_user' => (string)($vars['DB_USER'] ?? ''),
	'db_username' => (string)($vars['DB_USERNAME'] ?? ''),
	'db_password' => (string)($vars['DB_PASSWORD'] ?? ''),
	'mysql_user' => (string)($vars['MYSQL_USER'] ?? ''),
	'mysql_database' => (string)($vars['MYSQL_DATABASE'] ?? ''),
	'mysql_password' => (string)($vars['MYSQL_PASSWORD'] ?? ''),
	'mysql_host' => (string)($vars['MYSQL_HOST'] ?? ''),
	'mysql_port' => (string)($vars['MYSQL_PORT'] ?? ''),
];
echo base64_encode(json_encode($out));
`
