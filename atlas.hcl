data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./swiftwave_service/db_models_loader",
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/15/dev"
  migration {
    dir = "file://swiftwave_service/db/migrations"
    format = golang-migrate
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}