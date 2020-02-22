package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB // 连接池对象

func initDB() (err error) {
	dsn := "root:xiaoyu0000@tcp(127.0.0.1:3306)/my_db"
	db, err = sql.Open("mysql", dsn) //不会校验用户名密码是否真确
	if err != nil {                  //dsn格式不真确的时候会报错
		fmt.Printf("dsn:%s invalid, err:%v\n", dsn, err)
		return
	}
	err = db.Ping()
	if err != nil {
		fmt.Printf("open %s failed, err:%v\n", dsn, err)
		return
	}
	db.SetMaxOpenConns(10) //设置数据库连接池的最大连接数
	db.SetMaxIdleConns(2)  // 设置数据库最大空闲连接数
	return
}

type user struct {
	id   int
	name string
	age  int
}

// 查询单条数据
func queryRow(id int) {
	var u user
	// 查询语句
	// sqlStr := "select id, name, age from user where id=1;"
	sqlStr := "select id, name, age from go_studyMysql where id=?;"
	// 执行sql
	// rowObj := db.QueryRow(sqlStr, 2) // 从连接池里拿一个连接出来去数据库查询单条数据
	// 拿到结果
	// rowObj.Scan(&u.id, &u.name, &u.age) //必须对rowObj对象调用Scan方法，因为该方法会释放连接
	db.QueryRow(sqlStr, id).Scan(&u.id, &u.name, &u.age)
	fmt.Printf("u:%v\n", u)
}

// 查询多条数据
func queryMore(n int) {
	sqlStr := "select id, name, age from go_studyMysql where id > ?;"
	rows, err := db.Query(sqlStr, n)
	if err != nil {
		fmt.Printf("exec %s query failed, err:%v\n", sqlStr, err)
		return
	}
	// 一定要关闭rows
	defer rows.Close() // 释放连接
	// 循环取值
	for rows.Next() {
		var u user
		err := rows.Scan(&u.id, &u.name, &u.age)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return
		}
		fmt.Printf("u:%v\n", u)
	}
}

// 插入数据
func insert(name string, age int) {
	sqlStr := `insert into go_studyMysql(name, age) values(?, ?)`
	ret, err := db.Exec(sqlStr, name, age)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	// 如果是插入数据操作，能够拿到插入数据的id值
	id, err := ret.LastInsertId()
	if err != nil {
		fmt.Printf("get last id failed, err:%v\n", err)
		return
	}
	fmt.Println("插入成功 id=", id)
}

// 更新数据
func updateRow(newAge, id int) {
	sqlStr := `update go_studyMysql set age=? where id = ?`
	ret, err := db.Exec(sqlStr, newAge, id)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return
	}
	n, err := ret.RowsAffected() //操作影响行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
	fmt.Printf("update success, affected rows:%d\n", n)
}

// 删除数据
func deleteRow(id int) {
	sqlStr := `delete from go_studyMysql where id=?`
	ret, err := db.Exec(sqlStr, id)
	if err != nil {
		fmt.Printf("delete failed, err%v\n", err)
		return
	}
	n, err := ret.RowsAffected()
	if err != nil {
		fmt.Printf("get id failed, err:%v\n", err)
		return
	}
	fmt.Printf("删除了%d行的数据", n)
}

// 预处理插入多条数据
func prepareInsert() {
	sqlStr := `insert into go_studyMysql(name, age) values(?, ?);`
	stmt, err := db.Prepare(sqlStr) // 吧SQL语句先发给MySql预处理一下
	if err != nil {
		fmt.Printf("prepare failed, err:%v\n", err)
		return
	}
	defer stmt.Close()
	// 后续只需要拿到stmt去执行操作
	var m = map[string]int{
		"史豹虎": 40,
		"任俊丽": 40,
	}
	for k, v := range m {
		stmt.Exec(k, v) // 传值
	}

}

// 事务
func transaction() {
	// 开启事务
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("begin failed, err:%v\n", err)
		return
	}
	// 执行多个sql操作
	sqlStr1 := "update go_studyMysql set age=age-2 where id=4;"
	sqlStr2 := "update go_studyMysql set age=age+2 where id=1;"

	// 执行sql1
	_, err = tx.Exec(sqlStr1)
	if err != nil {
		// 要回滚
		tx.Rollback()
		fmt.Println("执行sql1出错")
		return
	}

	// 执行sql2
	_, err = tx.Exec(sqlStr2)
	if err != nil {
		// 要回滚
		tx.Rollback()
		fmt.Println("执行sql2出错")
		return
	}

	// 上面俩个sql都执行成功，就提交本次事务
	err = tx.Commit()
	if err != nil {
		// 要回滚
		tx.Rollback()
		fmt.Println("提交出错")
		return
	}
	fmt.Println("事务执行成功")
}

// sql注入
func sqlInject(name string) {
	// 自己拼接SQL语句的字符串
	sqlStr := fmt.Sprintf("select id, name, age from go_studyMysql where name='%s'", name)
	fmt.Printf("SQL:%s\n", sqlStr)
	rows, err := db.Query(sqlStr)
	if err != nil {
		fmt.Printf("exec %s query failed, err:%v\n", sqlStr, err)
		return
	}
	// 一定要关闭rows
	defer rows.Close() // 释放连接
	// 循环取值
	for rows.Next() {
		var u user
		err := rows.Scan(&u.id, &u.name, &u.age)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return
		}
		fmt.Printf("u:%v\n", u)
	}
}

func main() {
	err := initDB()
	if err != nil {
		fmt.Printf("initDB failed, err:%v\n", err)
	}
	fmt.Println("连接数据库成功！")
	// queryMore(0) // 查询全部
	// insert("任用强", 30)
	// updateRow(20, 1)
	// deleteRow(3)
	// prepareInsert()
	// transaction()

	//SQL注入示例
	// sqlInject("任用强")
	// sqlInject("xxx' or 1=1 #")
	sqlInject("xxx' union select * from go_studyMysql #")
}
