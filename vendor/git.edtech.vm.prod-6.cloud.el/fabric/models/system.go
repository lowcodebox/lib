package models

import (
	"fmt"
	"time"
)

type RollingStrategy string

const (
	KILL_FIRST_ROLL = "kill_first"
	KILL_LAST_ROLL  = "kill_last"
)

// системный конфиг с общей структурой для всех сервисов
type Config struct {
	Project    string `envconfig:"PROJECT" default:""`
	ProjectKey string `envconfig:"PROJECT_KEY" default:""`
	Service    string `envconfig:"SERVICE" default:""` // имя сервиса например "oauth3"
	// Name используется в построение пути к сервису Path = {Project}/{Name}
	// для одного и того же сервиса (пр. app) могут быть разные названия (app_ch, app_eng)
	Name        string `envconfig:"NAME" default:"" description:"название бинарника, пр. oauth3v2"`
	Environment string `envconfig:"ENVIRONMENT" default:"dev"`

	Port                 string        `envconfig:"PORT" default:"0"`
	PortInterval         string        `envconfig:"PORT_INTERVAL" default:"8010:8100"`
	ProxyHost            string        `envconfig:"PROXY_HOST" default:"http://127.0.0.1/"`
	ProxyMaxCountRetries Int           `envconfig:"PROXY_MAX_COUNT_RETRIES" default:"12"`
	ProxyTimeRetries     time.Duration `envconfig:"PROXY_TIME_RETRIES" default:"5s"`

	Uid        string `envconfig:"UID" default:""`        // uid сервиса
	ReplicaID  string `envconfig:"REPLICA_ID" default:""` // id реплики назначается при старте сервиса
	Version    string `envconfig:"VERSION" default:""`
	HashCommit string `envconfig:"HASH_COMMIT" default:""`

	DC      string `envconfig:"DC" default:""`
	Cluster string `envconfig:"CLUSTER" default:""`

	// Infra settings
	Strategy RollingStrategy `json:"straregy" default:"kill_last"`

	// Logger
	LogsLevel string `envconfig:"LOGS_LEVEL" default:"debug"`
	// LOGBOX
	LogboxEndpoint       string        `envconfig:"LOGBOX_ENDPOINT" default:"127.0.0.1:8999"`
	LogboxAccessKeyId    string        `envconfig:"LOGBOX_ACCESS_KEY_ID" default:""`
	LogboxSecretKey      string        `envconfig:"LOGBOX_SECRET_KEY" default:""`
	LogboxRequestTimeout time.Duration `envconnfig:"LOGBOX_REQUEST_TIMEOUT" default:"300ms"`
	// LOGBOX-client CircuitBreaker
	CbMaxRequestsLogbox uint32        `envconfig:"CB_MAX_REQUESTS_LOGBOX" default:"3" description:"максимальное количество запросов, которые могут пройти, когда автоматический выключатель находится в полуразомкнутом состоянии"`
	CbTimeoutLogbox     time.Duration `envconfig:"CB_TIMEOUT_LOGBOX" default:"5s" description:"период разомкнутого состояния, после которого выключатель переходит в полуразомкнутое состояние"`
	CbIntervalLogbox    time.Duration `envconfig:"CB_INTERVAL_LOGBOX" default:"5s" description:"циклический период замкнутого состояния автоматического выключателя для сброса внутренних счетчиков"`

	ExtendedLogs Bool `envconfig:"EXTENDED_LOGS" default:"false" description:"Логи сервиса включают параметры реквестов"`
	LogsExport   Bool `envconfig:"LOGS_EXPORT" default:"false" description:"Экспортировать логи сервиса"`

	UseAuthMiddleware Bool `envconfig:"USE_AUTH_MIDDLEWARE" default:"true" description:"Использовать аутентификационную middleware"`

	ReadTimeout  time.Duration `envconnfig:"READ_TIMEOUT" default:"10s"`
	WriteTimeout time.Duration `envconnfig:"WRITE_TIMEOUT" default:"10s"`

	// Registry
	Registry       string `envconfig:"REGISTRY" default:"http://git.edtech.vm.prod-6.cloud.el:8000"`
	RegistryDomain string `envconfig:"REGISTRY_DOMAIN" default:"controller/registry"`

	// Mesh settings
	// путь к меш-карте, которую составляет агент
	MeshMapPath string `envconnfig:"WRITE_TIMEOUT" default:"/opt/lowcodebox/service_map.json"`

	// Canary
	CanarySocketPath string `envconnfig:"CANARY_SOCKET_PATH" default:"/opt/lowcodebox/canary.sock"`

	// Extensions
	ExtensionsHost string `envconnfig:"EXTENSIONS_HOST" default:"localhost:8998"`
}

func (c *Config) Domain() string {
	return fmt.Sprintf("%s/%s", c.Project, c.Name)
}

type VFSConfig struct {
	VfsBucket         string `envconfig:"VFS_BUCKET" default:""`
	VfsKind           string `envconfig:"VFS_KIND" default:"s3"`
	VfsEndpoint       string `envconfig:"VFS_ENDPOINT" default:"http://127.0.0.1:9000"`
	VfsAccessKeyID    string `envconfig:"VFS_ACCESS_KEY_ID" default:"minioadmin"`
	VfsSecretKey      string `envconfig:"VFS_SECRET_KEY" default:"minioadmin"`
	VfsRegion         string `envconfig:"VFS_REGION" default:""`
	VfsComma          string `envconfig:"VFS_COMMA" default:""`
	VfsCertCA         string `envconfig:"VFS_CERT_CA" default:"" description:"CA-сертификат"`
	VfsCAFile         string `envconfig:"VFS_CA_FILE" default:"" description:"Файл CA-сертификата"`
	VfsCDNAccessKeyID string `envconfig:"VFS_CDN_ACCESS_KEY_ID" default:""`
	VfsCDNSecretKey   string `envconfig:"VFS_CDN_SECRET_KEY" default:""`
	VfsUploadPath     string `envconfig:"VFS_UPLOAD_PATH" default:"" description:"Путь, который по умолчанию подставляется ко всем загружаемым/читаемым файлам"`
}

type OrmTopicNames struct {
	Create       string `envconfig:"KAFKA_TOPIC_CREATE" default:"orm-create" description:"Топик с запросами на создание объектов"`
	Update       string `envconfig:"KAFKA_TOPIC_UPDATE" default:"orm-update" description:"Топик с запросами на обновление объектов"`
	BlockUnblock string `envconfig:"KAFKA_TOPIC_BLOCKUNBLOCK" default:"orm-blockunblock" description:"Топик с запросами на \"блокировку-разблокировку\" объектов"`
	LinkAdd      string `envconfig:"KAFKA_TOPIC_LINKADD" default:"orm-linkadd" description:"Топик с запросами на добавления ссылок в объекте"`
	LinkDelete   string `envconfig:"KAFKA_TOPIC_LINKDELETE" default:"orm-linkdekete" description:"Топик с запросами на удаление ссылок в объекте"`
	Commit       string `envconfig:"KAFKA_TOPIC_COMMIT" default:"orm-commit" description:"Топик с запросами на слитие веток"`
	Revert       string `envconfig:"KAFKA_TOPIC_REVERT" default:"orm-revert" description:"Топик с запросами на отмену слияния"`
}

type AnalyticsTopicNames struct {
	Set        string `envconfig:"KAFKA_TOPIC_SET" default:"analytics-set"`
	QueryAsync string `envconfig:"KAFKA_TOPIC_QUERY_ASYNC" default:"analytics-query-async"`
}
