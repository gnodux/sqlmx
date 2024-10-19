SELECT {{allColumns .Columns}}
FROM {{n .TableName}}
WHERE {{n .PrimaryKey.ColumnName}} IN (?)
{{if .TenantKey}}
AND {{.TenantKey.ColumnName}}=?
{{end}}