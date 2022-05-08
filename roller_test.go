package sqlroller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/blastrain/vitess-sqlparser/sqlparser"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func jsonString(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func Test1(_ *testing.T) {
	st := time.Now()
	stmt, err := sqlparser.Parse("select * from user_items where user_id=:x8 order by created_at limit ? offset 10")
	if err != nil {
		panic(err)
	}
	fmt.Println("stmt:", jsonString(stmt))
	q := sqlparser.String(stmt)
	fmt.Println("q:", q)
	sel := stmt.(*sqlparser.Select)
	fmt.Println("where:", jsonString(sel.Where))
	stmt2, err := sqlparser.Parse("select 1 from x where name=?")
	if err != nil {
		panic(err)
	}
	sel2 := stmt2.(*sqlparser.Select)
	fmt.Println("sel2:", jsonString(sel2.Where))
	//sel.AddWhere(sel2.Where.Expr)
	sel.Where.Expr = &sqlparser.OrExpr{
		Left:  sel.Where.Expr,
		Right: sel2.Where.Expr,
	}
	sel.SetLimit(&sqlparser.Limit{
		Offset:   sqlparser.NewIntVal([]byte("0")),
		Rowcount: sqlparser.NewValArg([]byte("?")),
	})
	fmt.Println("new where:", sqlparser.String(sel))
	fmt.Println("took:", time.Since(st))
}

func Test2(_ *testing.T) {
	q, args := R(`select id, en from feed where status=1`).And("user_id=?", 1).String()
	fmt.Println("q:", q, args)
}

func Test3(_ *testing.T) {
	q, args := R(`select g.en, count(distinct p.lang) n from package p 
inner join game g on g.id=p.game_id
where g.visible=1
group by g.id`).And("g.genre=?", 1).Limit(0, 10).String()
	fmt.Println("q:", q, args)
}

func Test4(_ *testing.T) {
	q, args := R(`select id, en from feed where status=1`).And("user_id=?", 1).
		OrderBy("id desc").Limit(0, 10).String()
	fmt.Println(q, args)
}

func Test5(_ *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/gamepal")
	if err != nil {
		panic(err)
	}
	q, args := R(`select id, en, translated from game1`).And("translated=?", 1).
		OrderBy("id desc").Limit(5, 10).String()
	fmt.Println("q:", q, args)
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
}
