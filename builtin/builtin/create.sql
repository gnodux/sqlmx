INSERT INTO {{n .TableName}}
({{columns .Columns}})
VALUES
({{args .Columns}})