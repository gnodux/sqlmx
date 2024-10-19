/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

SELECT *
FROM `user`
WHERE `id` IN ({{list .ids}})