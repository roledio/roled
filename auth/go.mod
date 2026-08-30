module github.com/roledio/roled/auth

go 1.26

require (
	github.com/aws/aws-sdk-go-v2 v1.41.1
	github.com/aws/aws-sdk-go-v2/config v1.32.7
	github.com/aws/aws-sdk-go-v2/credentials v1.19.7
	github.com/aws/aws-sdk-go-v2/service/s3 v1.96.0
	github.com/dustin/go-humanize v1.0.1
	github.com/ggwhite/go-masker/v2 v2.2.0
	github.com/gofiber/contrib/v3/jwt v1.2.1
	github.com/gofiber/contrib/v3/newrelic v1.1.9
	github.com/gofiber/contrib/v3/zap v1.0.10
	github.com/gofiber/fiber/v3 v3.5.0
	github.com/gofiber/template/html/v3 v3.0.2
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/gookit/goutil v0.7.3
	github.com/jinzhu/copier v0.4.0
	github.com/matoous/go-nanoid/v2 v2.1.1-0.20251203170756-2ab893bb7af4
	github.com/newrelic/go-agent/v3 v3.44.2
	github.com/redis/go-redis/v9 v9.22.0
	github.com/stretchr/testify v1.12.0
	github.com/tidwall/gjson v1.18.0
	gorm.io/driver/mysql v1.6.0
	gorm.io/driver/postgres v1.6.2
	gorm.io/gorm v1.31.2
	moul.io/zapgorm2 v1.3.0
)

require (
	github.com/MicahParks/keyfunc/v2 v2.1.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.4 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.6 // indirect
	github.com/aws/smithy-go v1.24.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gofiber/schema v1.8.4 // indirect
	github.com/gofiber/template/v2 v2.1.0 // indirect
	github.com/gofiber/utils/v2 v2.4.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tinylib/msgp v1.6.4 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Masterminds/squirrel v1.5.4
	github.com/andybalholm/brotli v1.2.2 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.1
	github.com/go-playground/validator/v10 v10.22.0
	github.com/go-sql-driver/mysql v1.9.3
	github.com/gofiber/storage/redis/v3 v3.5.2
	github.com/google/uuid v1.6.0
	github.com/govalues/decimal v0.1.29
	github.com/grokify/html-strip-tags-go v0.1.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/karrick/tparse/v2 v2.8.2
	github.com/klauspost/compress v1.19.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.10.9
	github.com/lithammer/shortuuid/v4 v4.2.0
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/newrelic/go-agent/v3/integrations/logcontext-v2/logWriter v1.0.3
	github.com/newrelic/go-agent/v3/integrations/logcontext-v2/nrwriter v1.0.2 // indirect
	github.com/newrelic/go-agent/v3/integrations/nrmysql v1.2.2
	github.com/newrelic/go-agent/v3/integrations/nrpq v1.1.1
	github.com/newrelic/go-agent/v3/integrations/nrredis-v9 v1.1.2
	github.com/pressly/goose/v3 v3.21.1
	github.com/rocketlaunchr/anti-disposable-email v1.0.0
	github.com/samber/lo v1.53.0
	github.com/sethvargo/go-retry v0.2.4 // indirect
	github.com/shomali11/util v0.0.0-20220717175126-f0771b70947f
	github.com/simukti/sqldb-logger v0.0.0-20230108155151-646c1a075551
	github.com/spf13/viper v1.21.0
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.73.0 // indirect
	go.openly.dev/pointy v1.3.0
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0
	golang.org/x/crypto v0.55.0
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260818201246-1b0934165a6f // indirect
	google.golang.org/grpc v1.83.1 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

replace github.com/simukti/sqldb-logger => github.com/ecionio/sqldb-logger v0.0.0-20260223084849-cecddfabd2aa

replace google.golang.org/genproto => google.golang.org/genproto/googleapis/api v0.0.0-20251222181119-0a764e51fe1b

exclude google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1
