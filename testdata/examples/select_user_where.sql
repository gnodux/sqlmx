SELECT *
FROM `user`
WHERE 1 = 1 {{if ne .Name ""}}
AND `name` LIKE '%{{.Name}}%'
{{end}}