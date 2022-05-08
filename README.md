# sqlroller
Dynamically extend SQL statements naturally, to avoid ORM.

## Idea
For me, ORM is too complex. Usually, my purpose is only to do paging query and add some filter conditions dynamically.

Why not write directly on the SQL statement?

## Example

```sql
# raw sql
select id, en from feed where status=1
```

To add where condition, We can write like this:

```go
import "github.com/zii/sqlroller"

q, args := sqlroller.R(`select id, en from feed where status=1`).And("user_id=?", 1).String()
fmt.Println(q, args)
```

output:
```sql
select id, en from feed where status = 1 and user_id = ? [1]
```

Custom orderby and limit:

```go
q, args := sqlroller.R(`select id, en from feed where status=1`).And("user_id=?", 1).
	OrderBy("id desc").Limit(0, 10).String()
fmt.Println(q, args)
```

output:
```sql
select id, en from feed where status=1 and user_id=1 order by id desc limit 0,10 [1]
```

ORM is difficult to handle multi table joint queries:

```go
q, args := R(`select g.en, count(distinct p.lang) n from package p 
inner join game g on g.id=p.game_id
where g.visible=1
group by g.id`).And("g.genre=?", 1).Limit(0, 10).String()
fmt.Println("q:", q, args)
```

output:
```sql
select g.en, count(distinct p.lang) as n from package as p join game as g on g.id = p.game_id where g.visible = 1
and g.genre = ? group by g.id limit 0, 10 [1]
```

More practical usage:

```go
import (
    "github.com/zii/sqlroller"
    _ "github.com/go-sql-driver/mysql"	
)

db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/gamepal")
raise(err)
q, args := R(`select id, en, translated from game1`).And("translated=?", 1).
    OrderBy("id desc").Limit(5, 10).String()
rows, err := db.Query(q, args...)
raise(err)
defer rows.Close()
for rows.Next() {
    var id int
    var en string
    var tr bool
    err := rows.Scan(&id, &en, &tr)
    raise(err)
    fmt.Println("row:", id, en, tr)
}
```

## Dependency Library
github.com/blastrain/vitess-sqlparser/sqlparser

## License
Unlicense
