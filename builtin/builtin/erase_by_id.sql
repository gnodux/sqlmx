DELETE
FROM {{n .TableName}}
WHERE {{n .PrimaryKey.ColumnName}} = :{{.PrimaryKey.ColumnName}}
{{if .TenantKey}}
AND {{n .TenantKey.ColumnName}}=:{{.TenantKey.ColumnName}}
{{end}}