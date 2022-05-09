package sqlroller

import (
	"encoding/json"
	"fmt"
	"github.com/blastrain/vitess-sqlparser/sqlparser"
	"strconv"
	"strings"
	"unicode"
)

func raise(err error) {
	if err != nil {
		panic(err)
	}
}

type SqlRoller struct {
	sel            *sqlparser.Select
	arg_map        map[string]interface{} // {placeholder:arg}
	where_clauses  []string
	orderby_clause string
	limit          *sqlparser.Limit
	id             int // placeholder ID, like :v1, :x2
}

func (this *SqlRoller) readExpr(in string, args []interface{}) string {
	for i := 0; i < len(args); i++ {
		id := this.NextId()
		p := fmt.Sprintf(":x%d", id)
		last := in
		in = strings.Replace(in, "?", p, 1)
		if in == last {
			break
		}
		this.arg_map[p] = args[i]
	}
	return in
}

func (this *SqlRoller) And(in string, args ...interface{}) *SqlRoller {
	in = this.readExpr(in, args)
	this.where_clauses = append(this.where_clauses, in)
	return this
}

func (this *SqlRoller) OrderBy(in string, args ...interface{}) *SqlRoller {
	in = this.readExpr(in, args)
	this.orderby_clause = in
	return this
}

func (this *SqlRoller) select2() *sqlparser.Select {
	where := strings.Join(this.where_clauses, " and ")
	q := "select * from x "
	if where != "" {
		q += "where " + where
	}
	if this.orderby_clause != "" {
		q += " order by " + this.orderby_clause
	}
	stmt, err := sqlparser.Parse(q)
	raise(err)
	sel := stmt.(*sqlparser.Select)
	return sel
}

func (this *SqlRoller) Limit(offset, size int) *SqlRoller {
	offset_s := strconv.Itoa(offset)
	size_s := strconv.Itoa(size)
	l := &sqlparser.Limit{
		Offset:   sqlparser.NewIntVal([]byte(offset_s)),
		Rowcount: sqlparser.NewIntVal([]byte(size_s)),
	}
	this.limit = l
	return this
}

func (this *SqlRoller) NextId() int {
	this.id++
	return this.id
}

func (this *SqlRoller) countQuestionMark(sql string) int {
	return strings.Count(sql, "?")
}

// initialize raw query and args
func (this *SqlRoller) init(sql string, args ...interface{}) {
	stmt, err := sqlparser.Parse(sql)
	raise(err)
	sel := stmt.(*sqlparser.Select)
	this.sel = sel
	n := this.countQuestionMark(sql)
	if len(args) < n {
		panic("number of args less than placeholders")
	}
	for i := 0; i < n; i++ {
		k := fmt.Sprintf(":v%d", i+1)
		this.arg_map[k] = args[i]
	}
}

func (this *SqlRoller) findPlaceholder(q string) (int, string) {
	i := strings.Index(q, ":")
	if i < 0 {
		return -1, ""
	}
	k := i
	n := len(q)
	i++
	if i >= n {
		return -1, ""
	}
	if q[i] != 'v' && q[i] != 'x' {
		return -1, ""
	}
	i++
	for ; i < n; i++ {
		if q[i] == ' ' || q[i] == ')' {
			break
		}
		if q[i] < '0' || q[i] > '9' {
			return -1, ""
		}
	}
	if i < k+2 {
		return -1, ""
	}
	return k, q[k:i]
}

func (this *SqlRoller) matchPrevToken(s, token string) bool {
	if token == "" {
		return false
	}
	tn := len(token)
	if len(s) < tn {
		return false
	}
	// skip space
	for i := len(s) - 1; i >= tn; i-- {
		if s[i] != '(' && !unicode.IsSpace(rune(s[i])) {
			break
		}
		s = s[:i]
	}
	if len(s) > tn && !unicode.IsSpace(rune(s[len(s)-tn-1])) {
		return false
	}
	return strings.ToLower(s[len(s)-tn:]) == strings.ToLower(token)
}

func (this *SqlRoller) encodeInArg(v interface{}) string {
	b, _ := json.Marshal(v)
	if len(b) < 2 {
		return string(b)
	}
	if b[0] == '[' {
		b = b[1:]
	}
	if b[len(b)-1] == ']' {
		b = b[:len(b)-1]
	}
	return string(b)
}

func (this *SqlRoller) String() (string, []interface{}) {
	var args []interface{}
	sel2 := this.select2()
	if sel2 != nil {
		if sel2.Where != nil {
			this.sel.AddWhere(sel2.Where.Expr)
		}
		if len(sel2.OrderBy) > 0 {
			this.sel.OrderBy = nil
			for _, o := range sel2.OrderBy {
				this.sel.AddOrder(o)
			}
		}
	}
	if this.limit != nil {
		this.sel.SetLimit(this.limit)
	}
	q := sqlparser.String(this.sel)
	builder := strings.Builder{}
	for q != "" {
		i, h := this.findPlaceholder(q)
		if i < 0 {
			builder.WriteString(q)
			break
		}
		builder.WriteString(q[:i])
		v := this.arg_map[h]
		in := this.matchPrevToken(q[:i], "in")
		if in {
			a := this.encodeInArg(v)
			builder.WriteString(a)
		} else {
			builder.WriteString("?")
			args = append(args, v)
		}
		q = q[i+len(h):]
	}
	return builder.String(), args
}

func R(raw string, args ...interface{}) *SqlRoller {
	r := new(SqlRoller)
	r.arg_map = make(map[string]interface{})
	r.id = 0
	r.init(raw, args...)
	return r
}
