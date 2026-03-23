package remote

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"govard/internal/engine"
)

type WordPressDBInfo struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type WordPressEnvironment struct {
	DB WordPressDBInfo
}

func ProbeWordPressEnvironment(remoteName string, remoteCfg engine.RemoteConfig) (WordPressEnvironment, error) {
	remoteCommand := buildMagentoRemoteCommand(remoteCfg.Path, `php -r `+shellQuoteRemote(wordpressDBProbePHP))
	encoded, err := runRemoteCapture(remoteName, remoteCfg, remoteCommand)
	if err != nil {
		return WordPressEnvironment{}, err
	}
	return decodeWordPressEnvironmentPayload(encoded)
}

func decodeWordPressEnvironmentPayload(encoded string) (WordPressEnvironment, error) {
	trimmed := strings.TrimSpace(encoded)
	if trimmed == "" {
		return WordPressEnvironment{}, fmt.Errorf("remote probe returned empty payload")
	}

	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return WordPressEnvironment{}, fmt.Errorf("decode remote probe payload: %w", err)
	}

	var payload struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
	}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return WordPressEnvironment{}, fmt.Errorf("parse remote probe payload: %w", err)
	}

	host, port := ParseMagentoDBHostPort(payload.Host)
	username := strings.TrimSpace(payload.Username)
	database := strings.TrimSpace(payload.DBName)
	if username == "" || database == "" {
		return WordPressEnvironment{}, fmt.Errorf("remote wp-config.php is missing DB_USER or DB_NAME")
	}

	return WordPressEnvironment{
		DB: WordPressDBInfo{
			Host:     host,
			Port:     port,
			Username: username,
			Password: payload.Password,
			Database: database,
		},
	}, nil
}

const wordpressDBProbePHP = `
$dbname = ""; $dbuser = ""; $dbpass = ""; $dbhost = "";
$content = @file_get_contents("wp-config.php");
if ($content) {
    if (preg_match("/define\s*\(\s*['\"]DB_NAME['\"]\s*,\s*['\"]([^'\"]+)['\"]\s*\)/", $content, $m)) $dbname = $m[1];
    if (preg_match("/define\s*\(\s*['\"]DB_USER['\"]\s*,\s*['\"]([^'\"]+)['\"]\s*\)/", $content, $m)) $dbuser = $m[1];
    if (preg_match("/define\s*\(\s*['\"]DB_PASSWORD['\"]\s*,\s*['\"]([^'\"]+)['\"]\s*\)/", $content, $m)) $dbpass = $m[1];
    if (preg_match("/define\s*\(\s*['\"]DB_HOST['\"]\s*,\s*['\"]([^'\"]+)['\"]\s*\)/", $content, $m)) $dbhost = $m[1];
    
    if (!$dbname || !$dbuser) {
        define('SHORTINIT', true);
        @include "wp-config.php";
        if (defined('DB_NAME')) $dbname = DB_NAME;
        if (defined('DB_USER')) $dbuser = DB_USER;
        if (defined('DB_PASSWORD')) $dbpass = DB_PASSWORD;
        if (defined('DB_HOST')) $dbhost = DB_HOST;
    }
}
$r = ["host" => (string)$dbhost, "username" => (string)$dbuser, "password" => (string)$dbpass, "dbname" => (string)$dbname];
echo base64_encode(json_encode($r));
`
