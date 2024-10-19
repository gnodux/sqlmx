UPDATE {{n .Meta.TableName}}
SET {{setArgs .Meta.Columns}}
WHERE {{n .Meta.PrimaryKey.ColumnName}}=:{{.Meta.PrimaryKey.ColumnName}}
{{if .Meta.TenantKey -}}
AND {{n .Meta.TenantKey.ColumnName}}=:{{.Meta.TenantKey.ColumnName}}
{{end}}