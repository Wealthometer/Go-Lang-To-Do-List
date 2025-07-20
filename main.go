package main
import "fmt"
import "unicode/utf8"

func main ()  {
	fmt.Println("Hello-World")
	var intNum int = 32767
	intNum = intNum + 1
	fmt.Println((intNum))

	var floatNum float64 = 12345678.9
	fmt.Println((floatNum))

	var intNum1 int = 14
	var intNum2 int = 3
	fmt.Println(intNum1/intNum2)
	fmt.Println(intNum1 % intNum2)

	var myString string = "Hello \n World"
	fmt.Println(myString)

	fmt.Println(utf8.RuneCountInString(myString))

	var myRune rune = 'a'
	fmt.Println(myRune)

	var myBoolean bool = false
	fmt.Println(myBoolean)

	var intNum3 int
	fmt.Println(intNum3)

	myVar := "Text"
	fmt.Println(myVar)

	var myVar2 string = "foo()"
	fmt.Println(myVar2)

	var1, var2 := 1, 2
	fmt.Println( var1 + var2)

	const myConst
}