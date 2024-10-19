UPDATE {{n .Meta.TableName}}
SET {{setArgs .Columns}}
WHERE {{n .Meta.PrimaryKey.ColumnName}}=:{{.Meta.PrimaryKey.ColumnName}}
{{- if .UseTenantId -}}
{{- if .Meta.TenantKey }}
 AND {{n .Meta.TenantKey.ColumnName}}=:{{.Meta.TenantKey.ColumnName}}
{{- end -}}
{{- end -}}