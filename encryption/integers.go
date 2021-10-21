package encryption

import (
	"github.com/yilisita/goNum"
	"math"
	"math/rand"
	"strconv"
	"time"
)

const(
	n = 10
	l = 100
	aBound  = 100
	bBound = 1000
	tBound = 100
)

var w = math.Pow(2, 45)

func KeySwitch(M, c goNum.Matrix) goNum.Matrix{
	cstar := GetBitVector(c)
	return goNum.DotPruduct(M, cstar)
}


// 将十进制数字转化为二进制字符串
func convertToBin(num int) string {
	s := ""
	var isNegative = false
	if num < 0{
		isNegative = true
		num = int(math.Abs(float64(num)))
	}
	//if num == 0 {
	//	return "0"
	//}

	// num /= 2 每次循环的时候 都将num除以2  再把结果赋值给 num
	for ;num > 0 ; num /= 2 {
		lsb := num % 2
		// strconv.Itoa() 将数字强制性转化为字符串
		s = strconv.Itoa(lsb) + s
	}
	var fill = ""
	for i := 0; i < l - len(s); i++{
		fill = "0" + fill
	}
	if isNegative{
		fill = "-" + fill[1:]
	}
	return fill + s
}


func reverse(str string) string {
	rs := []rune(str)
	len := len(rs)
	var tt []rune

	tt = make([]rune, 0)
	for i := 0; i < len; i++ {
		tt = append(tt, rs[len-i-1])
	}
	return string(tt[0:])
}


func GetRandomMatrix(row, col, bound int) goNum.Matrix{
	rand.Seed(time.Now().Unix())
	var A = goNum.ZeroMatrix(row,col)
	//var data = []float64
	var data = make([]float64, row*col)
	for i := 0; i < row * col; i++{
		data[i] = float64(rand.Intn(bound))
	}
	A.Data = data
	return A
}


func GetBitMatrix(s goNum.Matrix) goNum.Matrix{
	var powers = make([]float64, l)
	for i := 0; i < l; i++{
		powers[i] = math.Pow(2, float64(i))
	}
	var res = make([]float64, 0)
	for _, k := range s.Data{
		for _, j := range powers{
			res = append(res, k * j)
		}
	}
	var final = goNum.NewMatrix(s.Rows, s.Columns * l, res)
	return final
}


func GetBitVector(c goNum.Matrix) goNum.Matrix{
	var (
		res = make([]float64, 0)
		sign = 1
		s string
	)
	for _, i := range c.Data{
		s = convertToBin(int(i))
		if s[0] == '-' {
			sign = -1
			s = "0" + s[1:]
		}
		s = reverse(s)
		for _, j := range s{
			// 这里便利字符串有问题
			res = append(res, float64(int(j - 48) * sign))
		}
	}
	A := goNum.NewMatrix(l * c.Rows * c.Columns, 1, res)
	return A
}


func GetSecretKey(T goNum.Matrix) goNum.Matrix{
	I := goNum.IdentityE(T.Rows)
	var(
		ISlice = make([][]float64, 0)
		TSlice = make([][]float64, 0)
	)
	ISlice = goNum.Matrix2ToSlices(I)
	TSlice = goNum.Matrix2ToSlices(T)
	for i := 0; i < T.Rows; i++{
		ISlice[i] = append(ISlice[i], TSlice[i]...)
	}
	var res = make([]float64, 0)
	for _, s := range ISlice{
		for _, j := range s{
			res = append(res, j)
		}
	}
	A := goNum.NewMatrix(I.Rows, I.Columns + T.Columns, res)
	return A
}


func NearestInteger(x int) int{
	return int((float64(x) + (w + 1) / 2) / w)
}


func Decrypt(S, c goNum.Matrix) goNum.Matrix{
	// 他写的矩阵乘法有问题
	sc := goNum.DotPruduct(S, c)
	x := make([]float64, 0)
	var temp float64
	for _, i := range sc.Data{
		temp = float64(NearestInteger(int(i)))
		if temp < 0{
			temp = temp - 1
		}
		x = append(x, temp)
	}
	return goNum.NewMatrix(sc.Rows, 1, x)
}


func Encrypt(T, x goNum.Matrix) goNum.Matrix{
	I := goNum.IdentityE(x.Rows)
	var xSub = make([]float64, 0)
	xSub = append(xSub, x.Data...)
	var xS = goNum.NewMatrix(x.Rows, x.Columns, xSub)
	for i := 0; i < len(x.Data); i++{
		xS.Data[i] *= w
	}
	return KeySwitch(KeySwitchMatrix(I, T), xS)
}


func KeySwitchMatrix(S, T goNum.Matrix) goNum.Matrix{
	sStar := GetBitMatrix(S)
	A := GetRandomMatrix(T.Columns, sStar.Columns, aBound)
	E := GetRandomMatrix(sStar.Rows, sStar.Columns, bBound)
	up1 := goNum.AddMatrix(sStar, E)
	up2 := goNum.SubMatrix(up1, goNum.DotPruduct(T, A))
	ASLice := goNum.Matrix2ToSlices(A)
	USlice := goNum.Matrix2ToSlices(up2)
	for _, j := range ASLice{
		USlice = append(USlice, j)
	}
	var res = make([]float64, 0)
	for _, s := range USlice{
		for _, j := range s{
			res = append(res, j)
		}
	}
	return goNum.NewMatrix(E.Rows + A.Rows, A.Columns, res)
}


//func main(){
//	var(
//		x1 = GetRandomMatrix(n, 1, 100)
//		x2 = GetRandomMatrix(n, 1, 100)
//		T = GetRandomMatrix(n,n, tBound)
//		S = GetSecretKey(T)
//		c1 = Encrypt(T, x1)
//		c2 = Encrypt(T, x2)
//	)
//
//	//fmt.Println(x1)
//	//fmt.Println(p1)
//	//加法
//	fmt.Println("直接计算:", goNum.AddMatrix(x1, x2).Data)
//	fmt.Println("加密计算：", Decrypt(S, goNum.AddMatrix(c1, c2)).Data)


	//线性变换
