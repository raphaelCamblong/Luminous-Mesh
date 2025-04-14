module github.com/raphaelCamblong/Luminous-Mesh/control-plane/core

go 1.23.1

replace github.com/raphaelCamblong/Luminous-Mesh/control-plane/shared => ../shared

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/raphaelCamblong/Luminous-Mesh/control-plane/shared v0.0.0-00010101000000-000000000000
)
