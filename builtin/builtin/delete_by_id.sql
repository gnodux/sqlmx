{{if .LogicDeleteKey -}}
UPDATE {{n .TableName}}
SET {{n .LogicDeleteKey.ColumnName}} = 1
WHERE {{n .PrimaryKey.ColumnName}} = :{{.PrimaryKey.ColumnName}}
{{if .TenantKey}}
AND {{n .TenantKey.ColumnName}}=:{{.TenantKey.ColumnName}}
{{end}}
{{else}}
DELETE
FROM {{n .TableName}}
WHERE {{n .PrimaryKey.ColumnName}} = :{{.PrimaryKey.ColumnName}}
{{if .TenantKey}}
AND {{n .TenantKey.ColumnName}}=:{{.TenantKey.ColumnName}}
{{end}}
{{end}}