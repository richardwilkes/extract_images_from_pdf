module github.com/richardwilkes/extract_images_from_pdf

go 1.24.2

require (
	github.com/richardwilkes/toolbox v1.123.0
	github.com/unidoc/unipdf/v3 v3.68.0
	github.com/yookoala/realpath v1.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/pkg/term v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/unidoc/freetype v0.2.3 // indirect
	github.com/unidoc/pkcs7 v0.2.0 // indirect
	github.com/unidoc/timestamp v0.0.0-20200412005513-91597fd3793a // indirect
	github.com/unidoc/unitype v0.5.1 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/image v0.26.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/unidoc/unipdf/v3 => ../unipdf-override
