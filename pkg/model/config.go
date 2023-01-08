package model

type Config struct {
	HttpsOnly  string `envconfig:"HTTPS_ONLY" default:""`
	ProjectKey string `envconfig:"PROJECT_KEY" default:"LKHlhb899Y09olUi"`

	IntervalCleaner Duration `envconfig:"INTERVAL_CLEANER" default:"10m" description:"период очистки кеша сессий через запрос актуальных сессий в IAM"`

	// VFS
	VfsBucket      string `envconfig:"VFS_BUCKET" default:"buildbox"`
	VfsKind        string `envconfig:"VFS_KIND" default:"s3"`
	VfsEndpoint    string `envconfig:"VFS_ENDPOINT" default:"http://127.0.0.1:9000"`
	VfsAccessKeyID string `envconfig:"VFS_ACCESS_KEY_ID" default:"minioadmin"`
	VfsSecretKey   string `envconfig:"VFS_SECRET_KEY" default:"minioadmin"`
	VfsRegion      string `envconfig:"VFS_REGION" default:""`
	VfsComma       string `envconfig:"VFS_COMMA" default:""`

	// Cache
	TimeoutCacheGenerate Duration `envconfig:"TIMEOUT_CACHE_GENERATE" default:"3m" description:"интервал после которого будет реинициализировано обновление кеша для статуса updated"`

	// Config
	ConfigName          string `envconfig:"CONFIG_NAME" default:""`
	RootDir             string `envconfig:"ROOT_DIR" default:""`
	BuildModuleParallel Bool   `envconfig:"BUILD_MODULE_PARALLEL" default:"true"`
	CompileTemplates    Bool   `envconfig:"COMPILE_TEMPLATES" default:"false"`

	TimeoutBlockGenerate Duration `envconfig:"TIMEOUT_BLOCK_GENERATE" default:"10s" description:"интервал после которого будет завершена работа по генерации блока"`

	// Pay
	PayShopid        string `envconfig:"PAY_SHOPID" default:""`
	PaySecretKey     string `envconfig:"PAY_SECRET_KEY" default:""`
	PayRedirect      string `envconfig:"PAY_REDIRECT" default:""`
	PayTplOrders     string `envconfig:"PAY_TPL_ORDERS" default:""`
	PayErrorRedirect string `envconfig:"PAY_ERROR_REDIRECT" default:"list/page/errorpay"`
	MoneyGate        string `envconfig:"MONEY_GATE" default:"https://payment.yandex.net/api/v3/payments"`

	ClientPath string `envconfig:"CLIENT_PATH" default:""`
	UrlGui     string `envconfig:"URL_GUI" default:""`
	UrlProxy   string `envconfig:"URL_PROXY" default:""`
	UrlApi     string `envconfig:"URL_API" default:""`
	UrlIam     string `envconfig:"URL_IAM" default:""`
	UidService string `envconfig:"UID_SERVICE" default:""`

	PortInterval    string `envconfig:"PORT_INTERVAL" default:""`
	ProxyPointsrc   string `envconfig:"PROXY_POINTSRC" default:""`
	ProxyPointvalue string `envconfig:"PROXY_POINTVALUE" default:""`

	// Logger
	LogsDir               string   `envconfig:"LOGS_DIR" default:"logs"`
	LogsLevel             string   `envconfig:"LOGS_LEVEL" default:""`
	LogIntervalReload     Duration `envconfig:"LOG_INTERVAL_RELOAD" default:"10m" description:"интервал проверки необходимости пересозданния нового файла"`
	LogIntervalClearFiles Duration `envconfig:"LOG_INTERVAL_CLEAR_FILES" default:"30m" description:"интервал проверка на необходимость очистки старых логов"`
	LogPeriodSaveFiles    string   `envconfig:"LOG_PERION_SAVE_FILES" default:"0-1-0" description:"период хранения логов"`
	LogIntervalMetric     Duration `envconfig:"LOG_INTERVAL_METRIC" default:"10s" description:"период сохранения метрик в файл логирования"`

	TplUsers      Duration `envconfig:"TPL_USERS" default:""`
	TplRoles      Duration `envconfig:"TPL_ROLES" default:""`
	TplProfiles   string   `envconfig:"TPL_PROFILES" default:""`
	TplDatasource Duration `envconfig:"TPL_DATASOURCE" default:""`

	PK                              string `envconfig:"PK" default:""`
	ProcToleranceExcessLimitSession Float  `envconfig:"PROC_TOLERANCE_EXCESS_LIMIT_SESSION" default:"1.1"`
	Lang                            string `envconfig:"LANG" default:"RU"`
	Redirect_error                  string `envconfig:"REDIRECT_ERROR" default:"list/page/error"`
	Redirect_errorpay               string `envconfig:"REDIRECT_ERRORPAY" default:"list/page/errorpay"`

	MaxRequestBodySize Int      `envconfig:"MAX_REQUEST_BODY_SIZE" default:"10485760"`
	ReadTimeout        Duration `envconnfig:"READ_TIMEOUT" default:"10s"`
	WriteTimeout       Duration `envconnfig:"WRITE_TIMEOUT" default:"10s"`
	ReadBufferSize     Int      `envconfig:"READ_BUFFER_SIZE" default:"16384"`

	// Params from .cfg
	SlashDatecreate      string `envconfig:"SLASH_DATECREATE" default:""`
	SlashOwnerPointsrc   string `envconfig:"SLASH_OWNER_POINTSRC" default:""`
	SlashOwnerPointvalue string `envconfig:"SLASH_OWNER_POINTVALUE" default:""`
	SlashSemaforType     string `envconfig:"SLASH_SEMAFOR_TYPE" default:""`
	SlashTitle           string `envconfig:"SLASH_TITLE" default:""`

	AccessAdminPointsrc    string `envconfig:"ACCESS_ADMIN_POINTSRC" default:""`
	AccessAdminPointvalue  string `envconfig:"ACCESS_ADMIN_POINTVALUE" default:""`
	AccessDeletePointsrc   string `envconfig:"ACCESS_DELETE_POINTSRC" default:""`
	AccessDeletePointvalue string `envconfig:"ACCESS_DELETE_POINTVALUE" default:""`
	AccessReadPointsrc     string `envconfig:"ACCESS_READ_POINTSRC" default:""`
	AccessReadPointvalue   string `envconfig:"ACCESS_READ_POINTVALUE" default:""`
	AccessWritePointsrc    string `envconfig:"ACCESS_WRITE_POINTSRC" default:""`
	AccessWritePointvalue  string `envconfig:"ACCESS_WRITE_POINTVALUE" default:""`
	AppLevelLogsPointsrc   string `envconfig:"APP_LEVEL_LOGS_POINTSRC" default:""`
	AppLevelLogsPointvalue string `envconfig:"APP_LEVEL_LOGS_POINTVALUE" default:""`
	AppVersionPointsrc     string `envconfig:"APP_VERSION_POINTSRC" default:""`
	AppVersionPointvalue   string `envconfig:"APP_VERSION_POINTVALUE" default:""`

	BaseCache string `envconfig:"BASE_CACHE" default:""`

	Cache           string `envconfig:"CACHE" default:""`
	CachePointsrc   string `envconfig:"CACHE_POINTSRC" default:""`
	CachePointvalue string `envconfig:"CACHE_POINTVALUE" default:""`

	CopiesServiceapp string `envconfig:"COPIES_SERVICEAPP" default:""`

	DataSource  string `envconfig:"DATA_SOURCE" default:""`
	DataUid     string `envconfig:"DATA_UID" default:""`
	Description string `envconfig:"DESCRIPTION" default:""`
	Domain      string `envconfig:"DOMAIN" default:""`
	Driver      string `envconfig:"DRIVER" default:""`

	Error500 string `envconfig:"ERROR500" default:""`

	Logo string `envconfig:"LOGO" default:""`

	Metric    string `envconfig:"METRIC" default:""`
	Namespace string `envconfig:"NAMESPACE" default:""`

	PathTemplates string `envconfig:"PATH_TEMPLATES" default:""`
	Projectuid    string `envconfig:"PROJECTUID" default:""`
	PortApp       string `envconfig:"PORT_APP" default:""`

	ReplicasApp Int    `envconfig:"REPLICAS_APP" default:""`
	Robot       string `envconfig:"ROBOT" default:""`

	Signin    string `envconfig:"SIGNIN" default:""`
	SigninUrl string `envconfig:"SIGNIN_URL" default:""`

	Title                   string `envconfig:"TITLE" default:""`
	ToBuild                 string `envconfig:"TO_BUILD" default:""`
	ToUpdate                string `envconfig:"TO_UPDATE" default:""`
	TplAppBlocksPointsrc    string `envconfig:"TPL_APP_BLOCKS_POINTSRC" default:""`
	TplAppBlocksPointvalue  string `envconfig:"TPL_APP_BLOCKS_POINTVALUE" default:""`
	TplAppMaketsPointsrc    string `envconfig:"TPL_APP_MAKETS_POINTSRC" default:""`
	TplAppMaketsPointvalue  string `envconfig:"TPL_APP_MAKETS_POINTVALUE" default:""`
	TplAppModulesPointsrc   string `envconfig:"TPL_APP_MODULES_POINTSRC" default:""`
	TplAppModulesPointvalue string `envconfig:"TPL_APP_MODULES_POINTVALUE" default:""`
	TplAppPagesPointsrc     string `envconfig:"TPL_APP_PAGES_POINTSRC" default:""`
	TplAppPagesPointvalue   string `envconfig:"TPL_APP_PAGES_POINTVALUE" default:""`

	UrlFs string `envconfig:"URL_FS" default:""`

	Workingdir string `envconfig:"WORKINGDIR" default:""`
}
