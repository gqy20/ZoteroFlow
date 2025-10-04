package main

import "fmt"

// 这是一个故意格式错误的测试文件
func main(){
fmt.Println("测试pre-commit hook自动修复格式化问题")
var x = 1
var y = 2
if x>0{
fmt.Println("x是正数")
}
}