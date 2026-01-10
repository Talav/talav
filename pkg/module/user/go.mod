module github.com/talav/talav/pkg/module/userhttp

go 1.25.0

require (
	github.com/talav/talav/pkg/component/security v0.0.0-00010101000000-000000000000
	github.com/talav/talav/pkg/component/user v0.0.0-00010101000000-000000000000
	github.com/talav/talav/pkg/component/zorya v0.0.0-20260108152727-349eb6dbc95e
	go.uber.org/fx v1.24.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.11 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.29.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/talav/talav/pkg/component/mapstructure v0.0.0-20251212040909-717bc712a8cc // indirect
	github.com/talav/talav/pkg/component/negotiation v0.0.0-20251213015208-199315015cbe // indirect
	github.com/talav/talav/pkg/component/schema v0.0.0-20251213015208-199315015cbe // indirect
	github.com/talav/talav/pkg/component/tagparser v0.0.0-20251210172924-f671c53a0295 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)

replace (
	github.com/talav/talav/pkg/component/mapstructure => ../../component/mapstructure
	github.com/talav/talav/pkg/component/negotiation => ../../component/negotiation
	github.com/talav/talav/pkg/component/orm => ../../component/orm
	github.com/talav/talav/pkg/component/schema => ../../component/schema
	github.com/talav/talav/pkg/component/security => ../../component/security
	github.com/talav/talav/pkg/component/tagparser => ../../component/tagparser
	github.com/talav/talav/pkg/component/user => ../../component/user
	github.com/talav/talav/pkg/component/validator => ../../component/validator
	github.com/talav/talav/pkg/component/zorya => ../../component/zorya
)
