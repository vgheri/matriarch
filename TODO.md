PRIORITY TODO:

- [ ] Delete statement use case

NORMAL TODO:

- Insert statement use case
  - query_processor.Process:
    - [ ] Investigate how we can use either ExecParams either the low level connection to PGSQL by hijacking the connection
- [ ] add routine that listens to signals such as SIGTERM or SIGKILL and which kills the mock
- [ ] add connection pooling to main.go to manage frontend and backend connections

## Message flow between client and server

insert into company(id, name, age) values(1, 'Test ciccio', 24);
select _ from company;
update company set salary = 78000 where id = 1;
select _ from company where id = 1;
insert into orders(id, member_id, amount) values('8d007a96-d575-43d3-ab27-5a12c43f2963', '46e2e207-1b61-4099-9e99-f3e8630ebbd1', 456);
insert into company(id, name, age) values(5, 'Test Val', 24) RETURNING id;

F {"Type":"Query","String":"insert into company(id, name, age) values(3, 'Test ciccio', 24);"}
B {"Type":"CommandComplete","CommandTag":"INSERT 0 1"}
B {"Type":"ReadyForQuery","TxStatus":"I"}

F {"Type":"Query","String":"insert into company(id, name, age) values(1, 'Test ciccio', 24);"}
B {"Severity":"ERROR","Code":"23505","Message":"duplicate key value violates unique constraint \"company_pkey\"","Detail":"Key (id)=(1) already exists.","Hint":"","Position":0,"InternalPosition":0,"InternalQuery":"","Where":"","SchemaName":"public","TableName":"company","ColumnName":"","DataTypeName":"","ConstraintName":"company_pkey","File":"nbtinsert.c","Line":570,"Routine":"\_bt_check_unique","UnknownFields":{"86":"ERROR"}}
B {"Type":"ReadyForQuery","TxStatus":"I"}

F {"Type":"Query","String":"select \* from company;"}
B {"Type":"RowDescription","Fields":[{"Name":"id","TableOID":16386,"TableAttributeNumber":1,"DataTypeOID":23,"DataTypeSize":4,"TypeModifier":-1,"Format":0},{"Name":"name","TableOID":16386,"TableAttributeNumber":2,"DataTypeOID":25,"DataTypeSize":-1,"TypeModifier":-1,"Format":0},{"Name":"age","TableOID":16386,"TableAttributeNumber":3,"DataTypeOID":23,"DataTypeSize":4,"TypeModifier":-1,"Format":0},{"Name":"address","TableOID":16386,"TableAttributeNumber":4,"DataTypeOID":1042,"DataTypeSize":-1,"TypeModifier":54,"Format":0},{"Name":"salary","TableOID":16386,"TableAttributeNumber":5,"DataTypeOID":700,"DataTypeSize":4,"TypeModifier":-1,"Format":0}]}
B {"Type":"DataRow","Values":[{"text":"1"},{"text":"Valerio Gheri"},{"text":"38"},{"text":"311 chemin des combes "},{"text":"70000"}]}
B {"Type":"DataRow","Values":[{"text":"2"},{"text":"asdasda"},{"text":"23"},{"text":"ad2323addasd "},{"text":"23"}]}
B {"Type":"DataRow","Values":[{"text":"3"},{"text":"Test ciccio"},{"text":"24"},null,null]}
B {"Type":"CommandComplete","CommandTag":"SELECT 3"}
B {"Type":"ReadyForQuery","TxStatus":"I"}

F {"Type":"Query","String":"update company set salary = 78000 where id = 1;"}
B {"Type":"CommandComplete","CommandTag":"UPDATE 1"}
B {"Type":"ReadyForQuery","TxStatus":"I"}

F {"Type":"Query","String":"select \* from company where id = 1;"}
B {"Type":"RowDescription","Fields":[{"Name":"id","TableOID":16386,"TableAttributeNumber":1,"DataTypeOID":23,"DataTypeSize":4,"TypeModifier":-1,"Format":0},{"Name":"name","TableOID":16386,"TableAttributeNumber":2,"DataTypeOID":25,"DataTypeSize":-1,"TypeModifier":-1,"Format":0},{"Name":"age","TableOID":16386,"TableAttributeNumber":3,"DataTypeOID":23,"DataTypeSize":4,"TypeModifier":-1,"Format":0},{"Name":"address","TableOID":16386,"TableAttributeNumber":4,"DataTypeOID":1042,"DataTypeSize":-1,"TypeModifier":54,"Format":0},{"Name":"salary","TableOID":16386,"TableAttributeNumber":5,"DataTypeOID":700,"DataTypeSize":4,"TypeModifier":-1,"Format":0}]}
B {"Type":"DataRow","Values":[{"text":"1"},{"text":"Valerio Gheri"},{"text":"38"},{"text":"311 chemin des combes "},{"text":"78000"}]}
B {"Type":"CommandComplete","CommandTag":"SELECT 1"}
B {"Type":"ReadyForQuery","TxStatus":"I"}

F {"Type":"Query","String":"insert into orders(id, member_id, amount) values('8d007a96-d575-43d3-ab27-5a12c43f2963', '46e2e207-1b61-4099-9e99-f3e8630ebbd1', 456);"}
B {"Severity":"ERROR","Code":"42P01","Message":"relation \"orders\" does not exist","Detail":"","Hint":"","Position":13,"InternalPosition":0,"InternalQuery":"","Where":"","SchemaName":"","TableName":"","ColumnName":"","DataTypeName":"","ConstraintName":"","File":"parse_relation.c","Line":1194,"Routine":"parserOpenTable","UnknownFields":{"86":"ERROR"}}
B {"Type":"ReadyForQuery","TxStatus":"I"}

F {"Type":"Query","String":"insert into company(id, name, age) values(5, 'Test Val', 24) RETURNING id;"}
B {"Type":"RowDescription","Fields":[{"Name":"id","TableOID":16386,"TableAttributeNumber":1,"DataTypeOID":23,"DataTypeSize":4,"TypeModifier":-1,"Format":0}]}
B {"Type":"DataRow","Values":[{"text":"5"}]}
B {"Type":"CommandComplete","CommandTag":"INSERT 0 1"}
B {"Type":"ReadyForQuery","TxStatus":"I"}
