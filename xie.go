package xie

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/topxeq/goph"
	"github.com/topxeq/sqltk"
	"github.com/topxeq/tk"

	"github.com/mholt/archiver/v3"

	"github.com/kbinani/screenshot"
)

var VersionG string = "0.6.1"

var ShellModeG bool = false

type UndefinedStruct struct {
	int
}

func (o UndefinedStruct) String() string {
	return "未定义"
}

var Undefined UndefinedStruct = UndefinedStruct{0}

type XieHandler func(actionA string, dataA interface{}, paramsA ...interface{}) interface{}

// var TimeFormat = "2006-01-02 15:04:05"
// var TimeFormatMS = "2006-01-02 15:04:05.000"
// var TimeFormatMSCompact = "20060102150405.000"
// var TimeFormatCompact = "20060102150405"
// var TimeFormatCompact2 = "2006/01/02 15:04:05"
// var TimeFormatDateCompact = "20060102"
var ConstMapG map[string]interface{} = map[string]interface{}{
	"timeFormat":        "2006-01-02 15:04:05",
	"timeFormatCompact": "20060102150405",
}

// 指令集
// 关于结果参数：很多命令需要一个用于指定接收指令执行结果的参数，称作结果参数
// 结果参数一般都是一个变量，因此也称作结果变量
// 结果变量可以是$push（表示将结果压入堆栈中）、$drop（表示将结果丢弃）、$tmp（表示一个用于临时存储的全局变量）等预置全局变量
// 结果变量一般可以省略，此时表示将结果存入预置全局变量$tmp中（早期版本的谢语言默认是将结果压入堆栈，但从0.2.3版本之后均为存入$tmp）
// 当指令的参数个数可变时，结果参数不可省略，以免产生混淆
// 如果指令应返回结果，则当不提结果参数时，“第一个参数”一般指的是除结果参数外的第一个参数，余者类推

var InstrNameSet map[string]int = map[string]int{

	// internal & debug related
	"invalidInstr": 12, // 无效指令，用于内部表示有错误的指令

	"version": 100, // 获取当前谢语言版本（字符串），结果放入结果参数，如果不显式指定结果参数，结果默认将存入预置全局变量$tmp中（后面对于指令结果的处理对于大多数指令都是类似的，不再重复，特殊的情况是存在可变个数的参数的时候，结果变量不可省略）
	"版本":      100,

	"pass": 101, // 不做任何事情的指令
	"过":    101,

	"debug": 102, // 显示内部调试信息
	"调试":    102,

	"debugInfo": 103, // 获取调试信息
	"varInfo":   104, // 获取变量信息

	"help": 105, // 提供帮助信息

	"onError": 106, // 设置出错处理代码块，如有第一个参数，是一个标号，表示要跳转到的代码块位置，如无参数，表示清除（不设置任何错误处理代码块）

	"dumpf": 107, // 输出一定格式的变量调试信息，用法类似printf

	"defer": 109, // 延迟执行（主程序或函数退出时）

	"isUndef": 111, // 判断变量是否未被声明（定义），第一个结果参数可省略，第二个参数是要判断的变量
	"是否未定义":   111,
	"isDef":   112, // 判断变量是否已被声明（定义），第一个结果参数可省略，第二个参数是要判断的变量
	"是否已定义":   112,
	"isNil":   113, // 判断变量是否是nil，第一个结果参数可省略，第二个参数是要判断的变量

	"test": 121, // 内部测试用，测试两个数值是否相等

	"typeOf": 131, // 获取变量或数值类型（字符串格式），省略所有参数表示获取看栈值（不弹栈）的类型
	"类型":     131,

	"layer": 141, // 获取变量所处的层级（主函数层级为0，调用的第一个函数层级为1，再嵌套调用的为2，……）

	"loadCode": 151, // 载入字符串格式的谢语言代码到当前虚拟机中（加在最后），出错则返回TXERROR：开头的字符串说明原因
	"载入代码":     151,

	"len": 161, // 获取字符串、列表、映射等的长度，参数全省略表示取弹栈值
	"长度":  161,

	"fatalf": 170, // 类似pl输出信息后退出程序运行

	"goto": 180, // 无条件跳转到指定标号处
	"jmp":  180,
	"转到":   180,

	"exit": 199, // 退出程序运行
	"终止":   199,

	// var related
	"global": 201, // 声明全局变量
	"声明全局":   201,

	"var":  203, // 声明局部变量
	"声明变量": 203,

	"const": 205, // 获取预定义常量

	"ref": 210, // 获取变量的引用（取地址）
	"取引用": 210,

	"refNative": 211,

	"unref": 215, // 对引用进行解引用
	"解引用":   215,

	"assignRef": 218, // 根据引用进行赋值（将引用指向的变量赋值）
	"引用赋值":      218,

	// push/peek/pop related
	"push": 220, // 将数值压栈
	"入栈":   220,

	"push$": 221,

	"peek": 222, // 查看栈顶数值（不弹栈）
	"看栈":   222,
	// "peek$": 223,

	"pop":  224, // 弹出栈顶数值，结果参数如果省略相当于丢弃栈顶值
	"出栈":   224,
	"pop$": 225,

	// "peek*": 226, // from reg
	// "pop*":  227, // from reg

	// "pushInt": 231,
	// "pushInt$": 232,
	// "pushInt#": 233,
	// "pushInt*": 234,

	"clearStack": 240,

	// "pushLocal": 290,

	// reg related

	// "regInt":  310,
	// "regInt#": 312, // from number

	// shared sync map(cross-VM) related 全局同步映射相关（跨虚拟机）
	"getSharedMapItem":       301, // 获取全局映射变量，用法：getSharedMapItem $result key default，其中key是键名，default是可以省略的默认值（省略时如果没有值将返回undefined）
	"getSharedMapSize":       302,
	"tryGetSharedMapItem":    303,
	"tryGetSharedMapSize":    304,
	"setSharedMapItem":       311, // 设置全局映射变量，用法：setSharedMapItem $result key value
	"trySetSharedMapItem":    313,
	"deleteSharedMapItem":    321,
	"tryDeleteSharedMapItem": 323,
	"clearSharedMap":         331,
	"tryClearSharedMap":      333,

	"lockSharedMap":        341,
	"tryLockSharedMap":     342,
	"unlockSharedMap":      343,
	"readLockSharedMap":    346,
	"tryReadLockSharedMap": 347,
	"readUnlockSharedMap":  348,

	"quickClearSharedMap":      351,
	"quickGetSharedMapItem":    353,
	"quickGetSharedMap":        354,
	"quickSetSharedMapItem":    355,
	"quickDeleteSharedMapItem": 357,
	"quickSizeSharedMap":       359,

	// assign related
	"assign": 401, // 赋值（从局部变量到全局变量依次查找，如果没有则新建局部变量）
	"=":      401,
	"赋值":     401,

	// "assign$":   402,
	// "assignInt": 410,
	// "assignI":   411,

	"assignGlobal": 491, // 声明（如果未声明的话）并赋值一个全局变量

	"assignLocal": 492, // 声明（如果未声明的话）并赋值一个局部变量
	"局部赋值":        492,

	// if/else, switch related
	"if": 610, // 判断第一个参数（布尔类型，如果省略则表示取弹栈值）如果是true，则跳转到指定标号处
	"是则": 610,

	"ifNot": 611, // 判断第一个参数（布尔类型，如果省略则表示取弹栈值）如果是false，则跳转到指定标号处
	"否则":    611,

	// "if$":    618,
	// "if*":    619,
	// "ifNot$": 621,
	// "否则$":    621,

	"ifEval": 631, // 判断第一个参数（字符串类型）表示的表达式计算结果如果是true，则跳转到指定标号处
	"表达式是则":  631,

	"ifEmpty": 641, // 判断是否是空（值为undefined、nil、false、空字符串、小于等于0的整数或浮点数均会满足条件），是则跳转

	"ifEqual":    643, // 判断是否相等，是则跳转
	"ifNotEqual": 644, // 判断是否不等，是则跳转

	"ifErr":  651, // 判断是否是error对象或TXERROR字符串，是则跳转
	"ifErrX": 651, // 判断是否是error对象或TXERROR字符串，是则跳转

	// compare related
	"==": 701, // 判断两个数值是否相等，无参数时，比较两个弹栈值，结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值
	"等于": 701,

	"!=":  702, // 判断两个数值是否不等，无参数时，比较两个弹栈值，结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值
	"不等于": 702,

	"<":    703, // 判断两个数值是否是第一个数值小于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值
	"小于":   703,
	">":    704, // 判断两个数值是否是第一个数值大于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值
	"大于":   704,
	"<=":   705, // 判断两个数值是否是第一个数值小于等于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值
	"小于等于": 705,
	">=":   706, // 判断两个数值是否是第一个数值大于等于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值
	"大于等于": 706,

	// ">i":   710,
	// "<i":   720,
	// "整数小于": 720,
	// "<i$":  721,
	// "<i*":  722,

	"cmp": 790, // 比较两个数值，根据结果返回-1，0或1，分别表示小于、等于、大于，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值
	"比较":  790,

	// operator related
	"inc": 801, // 将某个整数变量的值加1，省略参数的话将操作弹栈值
	"加一":  801,
	"++":  801,

	// "inc$": 802,
	// "inc*": 803,

	"dec": 810, // 将某个整数变量的值减1，省略参数的话将操作弹栈值
	"减一":  810,
	"--":  810,

	// "dec$": 811,

	// "dec*":     812,
	// "intAdd":   820,
	// "整数加":      820,
	// "intAdd$":  821,
	// "整数加$":     821,
	// "intDiv":   831,
	// "floatAdd": 840,
	// "floatDiv": 848,

	"add": 901, // 两个数值相加，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值
	"+":   901,
	"加":   901,

	"sub": 902, // 两个数值相减，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值
	"-":   902,
	"减":   902,

	"mul": 903, // 两个数值相乘，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值
	"*":   903,
	"乘":   903,

	"div": 904, // 两个数值相除，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值
	"/":   904,
	"除":   904,

	"mod": 905, // 两个数值做取模计算，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值
	"%":   905,
	"取模":  905,

	"!": 930, // 取反操作符，对于布尔值取反，即true -> false，false -> true。对于其他数值，如果是未定义的变量（即Undefined），返回true，否则返回false

	"not": 931, // 逻辑非操作符，对于布尔值取反，即true -> false，false -> true，对于int、rune、byte等按位取反，即 0 -> 1， 1 -> 0

	"&&": 933, // 逻辑与操作符

	"||": 934, // 逻辑或操作符

	"?":          990, // 三元操作符，用法示例：? $result $a $s1 "abc"，表示判断变量$a中的布尔值，如果为true，则结果为$s1，否则结果值为字符串abc，结果值将放入结果变量result中，如果省略结果参数，结果值将会存入$tmp
	"ifThenElse": 990,

	"eval":      998, // 计算一个表达式
	"quickEval": 999, // 计算一个表达式（目前不支持函数）

	// func related 函数相关
	"call": 1010, // 调用指定标号处的函数
	"调用":   1010,

	"ret": 1020, // 函数内返回
	"返回":  1020,

	"callFunc": 1050, // 封装调用函数，一个参数是传入参数（压栈值）的个数（可省略），第二个参数是字符串类型的源代码
	"封装调用":     1050,

	"goFunc": 1060, // 并发调用函数，一个参数是传入参数（压栈值）的个数（可省略），第二个参数是字符串类型的源代码

	"fastCall": 1070, // 快速调用函数
	"快调":       1070,

	"fastRet": 1071, // 被快速调用的函数中返回
	"快回":      1071,

	"for": 1080, // for循环

	"range": 1085, // 遍历一个数字、字符串、数组、映射等，用法：range $list1 :label1，标号label1处应以continue来进行循环，break跳出，也可以用continueIf、breakIf后加条件来进行循环控制。对于整数，可以用range #i2 #i5 :label1的方式来遍历2至4。循环遍历中，对于整数、字符串和数组（切片）会将遍历值和序号依次压栈，循环体内需要弹栈这两个数值进行处理（注意弹栈时的顺序，先弹出序号，后弹出遍历值）；而对映射类对象，则依次将键值和键名压栈。

	// array/slice related 数组/切片相关
	"newList":  1101, // 新建一个数组，后接任意个元素作为数组的初始项
	"newArray": 1101,

	"addItem":      1110, //数组中添加项
	"addArrayItem": 1110,
	"addListItem":  1110,
	"增项":           1110,

	"addStrItem": 1111,

	"deleteItem":      1112, //数组中删除项
	"删项":              1112,
	"deleteArrayItem": 1112,
	"deleteListItem":  1112,

	"addItems":      1115, // 数组添加另一个数组的值
	"增多项":           1115,
	"addArrayItems": 1115,
	"addListItems":  1115,

	"getAnyItem": 1120,
	"任意类型取项":     1120,

	"setAnyItem": 1121,
	"任意类型置项":     1121,

	// "setItem": 1121,
	// "置项":      1121,
	// "getItemX": 1123,
	// "取项X":      1123,
	"getItem":      1123, // 从数组中取项，结果参数不可省略，之后第一个参数为数组对象，第2个为要获取第几项（从0开始），第3个参数可省略，为取不到项时的默认值，省略时返回undefined
	"getArrayItem": 1123,
	"取项":           1123,
	"[]":           1123,

	"setItem":      1124, // 修改数组中某一项的值
	"setArrayItem": 1124,
	"置项":           1124,
	// "setItemX": 1124,
	// "置项X":      1124,

	"slice": 1130, // 对列表（数组）切片，如果没有指定结果参数，将改变原来的变量。用法示例：slice $list4 $list3 #i1 #i5，将list3进行切片，截取序号1（包含）至序号5（不包含）之间的项，形成一个新的列表，放入变量list4中
	"切片":    1130,

	"rangeList": 1140, // 遍历数组
	"遍历列表":      1140,

	"rangeStrList": 1141, // 遍历字符串数组
	"遍历字符串列表":      1141,

	// control related 控制相关
	"continue": 1210, // 循环中继续
	"继续循环":     1210,

	"break": 1211, // 跳出循环（注意只能是用于for、range、rangeList、rangeMap等循环指令）
	"跳出循环":  1211,

	"continueIf": 1212, // 条件满足则继续循环
	"breakIf":    1213, // 条件满足则跳出循环

	// map related 映射相关
	"setMapItem": 1310, // 设置映射项，用法：setMapItem $map1 Name "李白"
	"置映射项":       1310,

	"deleteMapItem": 1312, // 删除映射项
	"删映射项":          1312,

	"getMapItem": 1320, // 获取指定序号的映射项，用法：getMapItem $result $map1 #i2，获取map1中的序号为2的项（即第3项），放入结果变量result中，如果有第4个参数则为默认值（没有找到映射项时使用的值），省略时将是undefined（可与全局内置变量$undefined比较）
	"取映射项":       1320,
	"{}":         1320,

	"rangeMap": 1340, // 遍历映射
	"遍历映射":     1340,

	// object related 对象相关

	"new": 1401, // 新建一个数据或对象，第一个参数为结果放入的变量（不可省略），第二个为字符串格式的数据类型或对象名，后面是可选的0-n个参数，目前支持byte、int等，注意一般获得的结果是引用（或指针）
	// "newVar": 1402, // （已废弃，用var代替）新建一个数据或对象，第一个参数为结果放入的变量（不可省略），第二个为字符串格式的数据类型或对象名，后面是可选的0-n个参数，目前支持byte、int等，注意一般获得的结果是不是引用（或指针）

	"method": 1403, // 对特定数据类型执行一定的方法，例如：method $result $str1 trimSet "ab"，将对一个字符串类型的变量str1去掉首尾的a和b字符，结果放入变量result中（注意，该结果参数不可省略，即使该方法没有返回数据，此时可以考虑用$drop）
	"mt":     1403,

	"member": 1405, // 获取特定数据类型的某个成员变量的值，例如：member $result $requestG "Method"，将获得http请求对象的Method属性值（GET、POST等），结果放入变量result中（注意，该结果参数不可省略，即使该方法没有返回数据，此时可以考虑用$drop）
	"mb":     1405,

	"mbSet": 1407, // 设置某个成员变量

	"newObj": 1410, // 新建一个对象，第一个参数为结果放入的变量，第二个为字符串格式的对象名，后面是可选的0-n个参数，目前支持string、any等

	"setObjValue": 1411, // 设置对象本体值
	"getObjValue": 1412, // 获取对象本体值

	"getMember": 1420, // 获取对象成员值
	"setMember": 1430, // 设置对象成员值

	"callObj": 1440, // 调用对象方法

	// string related 字符串相关
	"backQuote": 1501, // 获取反引号字符串
	"quote":     1503, // 将字符串进行转义（加上转义符，如“"”变成“\"”）
	"unquote":   1504, // 将字符串进行解转义

	"isEmpty": 1510, // 判断字符串是否为空
	"是否空串":    1510,

	"strAdd": 1520,

	"strSplit": 1530, // 按指定分割字符串分割字符串，结果参数不可省略，用法示例：strSplit $result $str1 "," 3，其中第3个参数可选（即可省略），表示结果列表最多的项数（例如为3时，将只按逗号分割成3个字符串的列表，后面的逗号将忽略；省略或为-1时将分割出全部）
	"分割字符串":    1530,

	"strReplace":   1540, // 字符串替换，用法示例：strReplace $result $str1 $find $replacement
	"strReplaceIn": 1543, // 字符串替换，可同时替换多个子串，用法示例：strReplace $result $str1 $find1 $replacement1 $find2 $replacement2

	"trim":    1550, // 字符串首尾去空白
	"strTrim": 1550,
	"去空白":     1550,

	"trimSet": 1551, // 字符串首尾去指定字符，除结果参数外第二个参数（字符串类型）指定去掉那些字符

	"trimSetLeft":  1553, // 字符串首去指定字符，除结果参数外第二个参数（字符串类型）指定去掉那些字符
	"trimSetRight": 1554, // 字符串尾去指定字符，除结果参数外第二个参数（字符串类型）指定去掉那些字符

	"trimPrefix": 1557, // 字符串首去指定字符串，除结果参数外第二个参数（字符串类型）指定去掉的子串，如果没有则返回原字符串
	"trimSuffix": 1558, // 字符串尾去指定字符串，除结果参数外第二个参数（字符串类型）指定去掉的子串，如果没有则返回原字符串

	"toUpper": 1561, // 字符串转为大写
	"toLower": 1562, // 字符串转为小写

	"strPad": 1563, // 字符串补零等填充操作，例如 strPad $result $strT #i5 -fill=0 -right=true，第一个参数是接收结果的字符串变量（不可省略），第二个是将要进行补零操作的字符串，第三个参数是要补齐到几位，默认填充字符串fill为字符串0，right（表示是否在右侧填充）为false（也可以直接写成-right），因此上例等同于strPad $result $strT #i5，如果fill字符串不止一个字符，最终补齐数量不会多于第二个参数指定的值，但有可能少

	"strContains":   1571, // 判断字符串是否包含某子串
	"strContainsIn": 1572, // 判断字符串是否包含任意一个子串，结果参数不可省略
	"strCount":      1573, // 计算字符串中子串的出现次数

	"strIn":  1581, // 判断字符串是否在一个字符串列表中出现，函数定义： 用法：strIn $result $originStr -it $sub1 "sub2"，第一个可变参数如果以“-”开头，将表示参数开关，-it表示忽略大小写，并且trim再比较（strA并不trim）
	"inStrs": 1581,

	"strStartsWith":   1582, // 判断字符串是否以某个子串开头
	"strStartsWithIn": 1583, // 判断字符串是否以某个子串开头（可以指定多个子串，符合任意一个则返回true），结果参数不可省略
	"strEndsWith":     1584, // 判断字符串是否以某个子串结束
	"strEndsWithIn":   1585, // 判断字符串是否以某个子串结束（可以指定多个子串，符合任意一个则返回true），结果参数不可省略

	// binary related 二进制数据相关
	"bytesToData": 1601,
	"dataToBytes": 1603,
	"bytesToHex":  1605,

	// time related 时间相关
	"now":  1910, // 获取当前时间
	"现在时间": 1910,

	"nowStrCompact": 1911, // 获取简化的当前时间字符串，如20220501080930
	"nowStr":        1912, // 获取当前时间字符串的正式表达
	"nowStrFormal":  1912, // 获取当前时间字符串的正式表达
	"nowTick":       1913, // 获取当前时间的Unix时间戳形式

	"nowUTC": 1918,

	"timeSub": 1921, // 时间进行相减操作
	"时间减":     1921,

	"timeToLocal":  1941,
	"timeToGlobal": 1942,

	"getTimeInfo": 1951,

	"timeAddDate": 1961,

	"formatTime": 1971, // 将时间格式化为字符串，用法：formatTime $endT $endT "2006-01-02 15:04:05"
	"timeToStr":  1971,

	// "toTime": 1981, // 参看10871， 将任意数值（可为字符串、时间戳字符串、时间等）转换为时间，字符串格式可以类似now、2006-01-02 15:04:05、20060102150405、2006-01-02 15:04:05.000，或10位或13位的Unix时间戳，可带可选参数-global、-defaultNow、-defaultErr、-defaultErrStr、-format=2006-01-02等，

	"timeToTick": 1991, // 时间转时间戳，时间可为toTime指令中参数的格式
	"tickToTime": 1993, // 时间戳转时间，时间戳可为整数或字符串

	// math related 数学相关
	"abs": 2100, // 取绝对值

	// command-line related 命令行相关
	"getParam": 10001, // 获取指定序号的命令行参数，结果参数外第一个参数为list或strList类型，第二个为整数，第三个为默认值（字符串类型），例：getParam $result $argsG 2 ""
	"获取参数":     10001,

	"getSwitch": 10002, // 获取命令行参数中指定的开关参数，结果参数外第一个参数为list或strList类型，第二个为类似“-code=”的字符串，第三个为默认值（字符串类型），例：getSwitch $result $argsG "-code=" ""，将获取命令行中-code=abc的“abc”部分。

	"ifSwitchExists": 10003, // 判断命令行参数中是否有指定的开关参数，结果参数外第一个参数为list或strList类型，第二个为类似“-verbose”的字符串，例：ifSwitchExists $result $argsG "-verbose"，根据命令行中是否含有-verbose返回布尔值true或false
	"switchExists":   10003,

	"ifSwitchNotExists": 10005,
	"switchNotExists":   10005,

	// print related 输出相关
	"pln": 10410, // 相当于其它语言的println函数
	"输出行": 10410,

	"plo":   10411, // 输出一个变量或数值的类型和值
	"输出值类型": 10411,

	"plos": 10412, // 输出多个变量或数值的类型和值

	"pl": 10420, // 相当于其它语言的printf函数再多输出一个换行符\n
	"输出": 10420,

	"plNow": 10422, // 相当于pl，之前多输出一个时间

	"plv": 10430, // 输出一个变量或数值的值的内部表达形式
	"输出值": 10430,

	"plvsr": 10433, // 输出多个变量或数值的值的内部表达形式，之间以换行间隔

	"plErr": 10440, // 输出一个error（表示错误的数据类型）信息

	"plErrStr": 10450, // 输出一个TXERROR字符串（表示错误的字符串，以TXERROR:开头，后面一般是错误原因描述）信息

	"spr": 10460, // 相当于其它语言的sprintf函数

	// scan/input related 输入相关
	"scanf":  10511, // 相当于其它语言的scanf函数
	"sscanf": 10512, // 相当于其它语言的sscanf函数

	// convert related 转换相关
	"convert": 10810, // 转换数值类型，例如 convert $a int
	"转换":      10810,

	// "convert$": 10811,

	"hex":        10821, // 16进制编码，对于数字高位在后
	"hexb":       10822, // 16进制编码，对于数字高位在前
	"unhex":      10823, // 16进制解码，结果是一个字节列表
	"hexToBytes": 10823,
	"toHex":      10824, // 任意数值16进制编码

	"toBool":  10831,
	"toByte":  10835,
	"toRune":  10837,
	"toInt":   10851, // 任意数值转整数，可带一个默认值（转换出错时返回该值），不带的话返回-1
	"toFloat": 10855,
	"toStr":   10861,
	"toTime":  10871, // 将任意数值（可为字符串、时间戳字符串、时间等）转换为时间，字符串格式可以类似now、2006-01-02 15:04:05、20060102150405、2006-01-02 15:04:05.000，或10位或13位的Unix时间戳，可带可选参数-global、-defaultNow、-defaultErr、-defaultErrStr、-format=2006-01-02等，
	"toAny":   10891,

	// err string(TXERROR:) related TXERROR错误字符串相关
	"isErrStr":  10910, // 判断是否是TXERROR字符串，用法：isErrStr $result $str1 $errMsg，第三个参数可选（结果参数不可省略），如有则当str1为TXERROR字符串时，会放入错误原因信息
	"getErrStr": 10921, // 获取TXERROR字符串中的错误原因信息（即TXERROR:后的内容）

	// "getErrStr$":  10922,
	"checkErrStr": 10931, // 判断是否是TXERROR字符串，是则退出程序运行

	// error related error相关
	"isErr":     10941, // 判断是否是error对象，结果参数不可省略，除结果参数外第一个参数是需要确定是否是error的对象，第二个可选变量是如果是error时，包含的错误描述信息
	"getErrMsg": 10942, // 获取error对象的错误信息

	"isErrX": 10943, // 同时判断是否是error对象或TXERROR字符串，用法：isErrX $result $err1 $errMsg，第三个参数可选（结果参数不可省略），如有会放入错误原因信息

	"checkErrX": 10945, // 检查后续变量或数值是否是error对象或TXERROR字符串，是则输出后中止

	// http request/response related HTTP请求相关
	"writeResp":       20110, // 写一个HTTP请求的响应
	"setRespHeader":   20111, // 设置一个HTTP请求的响应头，如setRespHeader $responseG "Content-Type" "text/json; charset=utf-8"
	"writeRespHeader": 20112, // 写一个HTTP请求的响应头状态，如writeRespHeader $responseG #i200
	"getReqHeader":    20113, // 获取一个HTTP请求的请求头信息
	"genJsonResp":     20114, // 生成一个JSON格式的响应字符，用法：genJsonResp $result $requestG "success" "Test passed!"，结果格式类似{"Status":"fail", "Value": "network timeout"}，其中Status字段表示响应处理结果状态，一般只有success和fail两种，分别表示成功和失败，如果失败，Value字段中为失败原因，如果成功，Value中为空或需要返回的信息
	"genResp":         20114,
	"serveFile":       20116,

	"newMux":          20121, // 新建一个HTTP请求处理路由对象，等同于 new mux
	"setMuxHandler":   20122, // 设置HTTP请求路由处理函数
	"setMuxStaticDir": 20123, // 设置静态WEB服务的目录，用法示例：setMuxStaticDir $muxT "/static/" "./scripts" ，设置处理路由“/static/”后的URL为静态资源服务，第1个参数为newMux指令创建的路由处理器对象变量，第2个参数是路由路径，第3个参数是对应的本地文件路径，例如：访问 http://127.0.0.1:8080/static/basic.xie，而当前目录是c:\tmp，那么实际上将获得c:\tmp\scripts\basic.xie

	"startHttpServer":  20151, // 启动http服务器，用法示例：startHttpServer $resultT ":80" $muxT ；可以后面加-go参数表示以线程方式启动，此时应注意主线程不要退出，否则服务器线程也会随之退出，可以用无限循环等方式保持运行
	"startHttpsServer": 20153, // 启动https(SSL)服务器，用法示例：startHttpsServer $resultT ":443" $muxT /root/server.crt /root/server.key -go

	// web related WEB相关
	"getWeb": 20210, // 发送一个HTTP网络请求，并获取响应结果（字符串格式），getWeb指令除了第一个参数必须是返回结果的变量，第二个参数是访问的URL，其他所有参数都是可选的，method可以是GET、POST等；encoding用于指定返回信息的编码形式，例如GB2312、GBK、UTF-8等；headers是一个JSON格式的字符串，表示需要加上的自定义的请求头内容键值对；参数中还可以有一个映射类型的变量或值，表示需要POST到服务器的参数，用法示例：getWeb $resultT "http://127.0.0.1:80/xms/xmsApi" -method=POST -encoding=UTF-8 -timeout=15 -headers=`{"Content-Type": "application/json"}` $mapT

	"downloadFile": 20220, // 下载文件

	"getResource":     20291, // 获取JQuery等常用的脚本或其他内置文本资源，一般用于服务器端提供内置的jquery等脚本嵌入，避免从互联网即时加载，第一个的参数是jquery.min.js等js文件的名称
	"getResourceList": 20293, // 获取可获取的资源名称列表

	// html related HTML相关
	"htmlToText": 20310, // 将HTML转换为字符串，用法示例：htmlToText $result $str1 "flat"，第3个参数开始是可选参数，表示HTML转文本时的选项

	// regex related 正则表达式相关
	// "regReplaceAllStr$": 20411,

	"regFindAll":   20421, // 获取正则表达式的所有匹配，用法示例：regFindAll $result $str1 $regex1 $group
	"regFind":      20423, // 获取正则表达式的第一个匹配，用法示例：regFind $result $str1 $regex1 $group
	"regFindFirst": 20423,
	"regFindIndex": 20425, // 获取正则表达式的第一个匹配的位置，返回一个整数数组，任意值为-1表示没有找到匹配，用法示例：regFindIndex $result $str1 $regex1

	"regMatch": 20431, // 判断字符串是否完全符合正则表达式，用法示例：regMatch $result "abcab" `a.*b`

	"regContains":   20441, // 判断字符串中是否包含符合正则表达式的子串
	"regContainsIn": 20443, // 判断字符串中是否包含符合任意一个正则表达式的子串
	"regCount":      20445, // 计算字符串中包含符合某个正则表达式的子串个数，用法示例：regCount $result $str1 $regex1

	"regSplit": 20451, // 用正则表达式分割字符串

	// system related 系统相关
	"sleep": 20501, // 睡眠指定的秒数（浮点数）
	"睡眠":    20501,

	"getClipText": 20511, // 获取剪贴板文本
	"获取剪贴板文本":     20511,

	"setClipText": 20512, // 设置剪贴板文本
	"设置剪贴板文本":     20512,

	"getEnv":    20521, // 获取环境变量
	"setEnv":    20522, // 设置环境变量
	"removeEnv": 20523, // 删除环境变量

	"systemCmd":       20601, // 执行一条系统命令，例如： systemCmd "cmd" "/k" "copy a.txt b.txt"
	"openWithDefault": 20603, // 用系统默认的方式打开一个文件，例如： openWithDefault "a.jpg"

	"getOSName": 20901, // 获取操作系统名称，如windows,linux,darwin等
	"getOsName": 20901,

	// file related 文件相关
	"loadText": 21101, // 从指定文件载入文本
	"载入文本":     21101,

	"saveText": 21103, // 保存文本到指定文件
	"保存文本":     21103,

	"loadBytes": 21105, // 从指定文件载入数据（字节列表）
	"载入数据":      21105,

	"saveBytes": 21106, // 保存数据（字节列表）到指定文件
	"保存数据":      21106,

	"loadBytesLimit": 21107, // 从指定文件载入数据（字节列表），不超过指定字节数

	"writeStr": 21201, // 写入字符串，可以向文件、字节数组、字符串等写入

	"createFile": 21501, // 新建文件，如果带-return参数，将在成功时返回FILE对象，失败时返回error对象，否则返回error对象，成功为nil，-overwrite有重复文件不会提示。如果需要指定文件标志位等，用openFile指令

	"openFile": 21503, // 打开文件，如果带-read参数，则为只读，-write参数可写，-create参数则无该文件时创建一个，-perm=0777可以指定文件权限标志位

	"closeFile": 21507, // 关闭文件

	"cmpBinFile": 21601, // 逐个字节比较二进制文件，用法： cmpBinFile $result $file1 $file2 -identical -verbose，如果带有-identical参数，则只比较文件异同（遇上第一个不同的字节就返回布尔值false，全相同则返回布尔值true），不带-identical参数时，将返回一个比较结果对象

	"fileExists":   21701, // 判断文件是否存在
	"ifFileExists": 21701,
	"isDir":        21702, // 判断是否是目录
	"getFileSize":  21703, // 获取文件大小
	"getFileInfo":  21705, // 获取文件信息，返回映射对象，参看genFileList命令

	"removeFile": 21801, // 删除文件
	"renameFile": 21803, // 重命名文件
	"copyFile":   21805, // 复制文件，用法 copyFile $result $fileName1 $fileName2，可带参数-force和-bufferSize=100000等

	// path related 路径相关

	"genFileList": 21901, // 生成目录中的文件列表，即获取指定目录下的符合条件的所有文件，例：getFileList $result `d:\tmp` "-recursive" "-pattern=*" "-exclusive=*.txt" "-withDir" "-verbose"，另有 -compact 参数将只给出Abs、Size、IsDir三项, -dirOnly参数将只列出目录（不包含文件），列表项对象内容类似：map[Abs:D:\tmpx\test1.gox Ext:.gox IsDir:false Mode:-rw-rw-rw- Name:test1.gox Path:test1.gox Size:353339 Time:20210928091734]
	"getFileList": 21901,

	"joinPath": 21902, // 合并文件路径，第一个参数是结果参数不可省略，第二个参数开始要合并的路径
	"合并路径":     21902, // 合并文件路径

	"getCurDir": 21905, // 获取当前工作路径
	"setCurDir": 21906, // 设置当前工作路径

	"getAppDir": 21907, // 获取应用路径（谢语言主程序路径）

	"extractFileName": 21910, // 从文件路径中获取文件名部分
	"extractFileExt":  21911, // 从文件路径中获取文件扩展名（后缀）部分
	"extractFileDir":  21912, // 从文件路径中获取文件目录（路径）部分
	"extractPathRel":  21915, // 从文件路径中获取文件相对路径（根据指定的根路径）

	"ensureMakeDirs": 21921,

	// console related 命令行相关
	"getInput":    22001, // 从命令行获取输入，第一个参数开始是提示字符串，可以类似printf加多个参数，用法：getInput $text1 "请输入%v个数字：" #i2
	"getPassword": 22003, // 从命令行获取密码输入（输入字符不显示），第一个参数是提示字符串

	// json related JSON相关
	"toJson": 22101, // 将对象编码为JSON字符串
	"toJSON": 22101,
	"JSON编码": 22101,

	"fromJson": 22102, // 将JSON字符串转换为对象
	"fromJSON": 22102,
	"JSON解码":   22102,

	// xml related XML相关
	"toXml": 22201, // 将对象编码为XML字符串
	"toXML": 22201,
	"XML编码": 22201,

	// random related 随机数相关

	"randomize": 23000, // 初始化随机种子

	"getRandomInt": 23001, // 获得一个随机整数，结果参数不可省略，此外带一个参数表示获取[0,参数1]之间的随机整数，带两个参数表示获取[参数1,参数2]之间的随机整数

	"getRandomFloat": 23003, // 获得一个介于[0, 1)之间的随机浮点数

	"genRandomStr": 23101, // 生成随机字符串，用法示例：genRandomStr $result -min=6 -max=8 -noUpper -noLower -noDigit -special -space -invalid，其中，除结果参数外所有参数均可选，-min用于设置最少生成字符个数，-max设置最多字符个数，-noUpper设置是否包含大写字母，-noLower设置是否包含小写字母，-noDigit设置是否包含数字，-special设置是否包含特殊字符，-space设置是否包含空格，-invalid设置是否包含一般意义上文件名中的非法字符，
	"getRandomStr": 23101,

	// encode/decode related 编码解码相关

	"md5": 24101, // 生成MD5编码

	"simpleEncode": 24201, // 简单编码，主要为了文件名和网址名不含非法字符
	"simpleDecode": 24203, // 简单编码的解码

	"urlEncode": 24301, // URL编码（http://www.aaa.com -> http%3A%2F%2Fwww.aaa.com）
	"urlDecode": 24303, // URL解码

	"base64Encode": 24401, // Base64编码，输入参数是[]byte字节数组或字符串
	"base64":       24401,
	"base64Decode": 24403, // Base64解码
	"unbase64":     24403,

	"htmlEncode": 24501, // HTML编码（&nbsp;等）
	"htmlDecode": 24503, // HTML解码

	"hexEncode": 24601, // 十六进制编码，仅针对字符串
	"hexDecode": 24603, // 十六进制解码，仅针对字符串

	"toUtf8": 24801, // 转换字符串或字节列表为UTF-8编码，结果参数不可省略，第一个参数为要转换的源字符串或字节列表，第二个参数表示原始编码（默认为GBK）
	"toUTF8": 24801,

	// encrypt/decrypt related 加密/解密相关

	"encryptText": 25101, // 用TXDEF方法加密字符串
	"decryptText": 25103, // 用TXDEF方法解密字符串

	"encryptData": 25201, // 用TXDEF方法加密数据（字节列表）
	"decryptData": 25203, // 用TXDEF方法解密数据（字节列表）

	// network relate 网络相关
	"getRandomPort": 26001, // 获取一个可用的socket端口（注意：获取后应尽快使用，否则仍有可能被占用）

	// database related 数据库相关
	"dbConnect": 32101, // 连接数据库，用法示例：dbConnect $db "sqlite3" `c:\tmpx\test.db`，或dbConnect $db "godror" `user/pass@129.0.9.11:1521/testdb`，结果参数外第一个参数为数据库驱动类型，目前支持sqlite3、mysql、mssql、godror（即oracle）等，第二个参数为连接字串
	"连接数据库":     32101,

	"dbClose": 32102, // 关闭数据库连接
	"关闭数据库":   32102,

	"dbQuery": 32103, // 在指定数据库连接上执行一个查询的SQL语句（一般是select等），返回数组，每行是映射（字段名：字段值），用法示例：dbQuery $rs $db $sql $arg1 $arg2 ...
	"查询数据库":   32103,

	"dbQueryRecs": 32104, // 在指定数据库连接上执行一个查询的SQL语句（一般是select等），返回二维数组（第一行为字段名），用法示例：dbQueryRecs $rs $db $sql $arg1 $arg2 ...
	"查询数据库记录":     32104,

	"dbExec": 32105, // 在指定数据库连接上执行一个有操作的SQL语句（一般是insert、update、delete等），用法示例：dbExec $rs $db $sql $arg1 $arg2 ...
	"执行数据库":  32105,

	// markdown related Markdown格式相关
	"renderMarkdown": 40001, // 将Markdown格式字符串渲染为HTML

	// image related 图像处理相关
	"newImage": 41001,

	"pngEncode": 41101, // 将图像保存为PNG文件或其他可写载体（如字符串），用法：pngEncode $errT $fileT $imgT

	"jpgEncode":  41103, // 将图像保存为JPG文件或其他可写载体（如字符串），用法：jpgEncode $errT $fileT $imgT -quality=90
	"jpegEncode": 41103,

	// screen related 屏幕相关
	"getActiveDisplayCount": 45001, // 获取活跃屏幕数量

	"getScreenResolution": 45011, // 获取指定屏幕的分辨率，用法：getScreenResolution $rectT -index=0 -format=rect，其中后面的参数均可选，index指定要获取的活跃屏幕号，主屏幕是0，format可以是rect、json或为空，参看例子代码getScreenInfo.xie

	"captureDisplay": 45021, // 屏幕截图，用法 captureDisplay $imgT 0，截取0号活跃屏幕（主屏幕）的全屏截图

	"captureScreen":     45023, // 屏幕区域截图，用法 captureScreenRect $imgT 100 100 640 480，截取主屏幕的坐标(100,100)为左上角，宽640，高480的区域截图，后面几个参数均可省略，默认截全屏
	"captureScreenRect": 45023,

	// token related 令牌相关
	"genToken":   50001, // 生成令牌，用法：genToken $result $appCode $userID $userRole -secret=abc，其中可选开关secret是加密秘钥，可省略
	"checkToken": 50003, // 检查令牌，用法：checkToken $result XXXXX -secret=abc -expire=2，其中expire是设置的超时秒数（默认为1440），如果成功，返回类似“appCode|userID|userRole|”的字符串；失败返回TXERROR字符串

	// run code related 运行代码相关
	"runCode": 60001, // 运行一段谢语言代码，在新的虚拟机中执行，除结果参数（不可省略）外，第一个参数是字符串类型的代码（必选，后面参数都是可选），第二个参数为任意类型的传入虚拟机的参数（虚拟机内通过inputG全局变量来获取该参数），后面的参数可以是一个字符串数组类型的变量或者多个字符串类型的变量，虚拟机内通过argsG（字符串数组）来对其进行访问。

	// line editor related 内置行文本编辑器有关
	"leClear": 70001, // 清空行文本编辑器缓冲区

	"leLoadStr":     70003, // 行文本编辑器缓冲区载入指定字符串内容，例：leLoadStr $textT "abc\nbbb\n结束"
	"leSetAll":      70003, // 等同于leLoadStr
	"leSaveStr":     70007, // 取出行文本编辑器缓冲区中内容，例：leSaveStr $result $s
	"leGetAll":      70007, // 等同于leSaveStr
	"leLoad":        70011, // 从文件中载入文本到行文本编辑器缓冲区中，例：leLoad $result `c:\test.txt`
	"leLoadFile":    70011, // 等同于leLoad
	"leLoadClip":    70012, // 从剪贴板中载入文本到行文本编辑器缓冲区中，例：leLoadClip $result
	"leLoadSSH":     70015, // 从SSH连接获取文本文件内容，用法：leLoadSSH 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -path=远端文件路径，结果变量不可省略，其他参数省略时将从之前获取内容的SSH连接中获取
	"leLoadUrl":     70016, // 从网址URL载入文本到行文本编辑器缓冲区中，例：leLoadUrl $result `http://example.com/abc.txt`
	"leSave":        70017, // 将行文本编辑器缓冲区中内容保存到文件中，例：leSave $result `c:\test.txt`
	"leSaveFile":    70017, // 等同于leSave
	"leSaveClip":    70023, // 将行文本编辑器缓冲区中内容保存到剪贴板中，例：leSaveClip $result
	"leSaveSSH":     70025, // 将编辑缓冲区内容保存到SSH连接中，如果不带参数，将从之前获取内容的SSH连接中获取，用法：leSaveSSH 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -path=远端文件路径。结果参数不可省略
	"leInsert":      70027, // 行文本编辑器缓冲区中的指定位置前插入指定内容，例：leInsert $result 3 "abc"
	"leInsertLine":  70027, // 等同于leInsert
	"leAppend":      70029, // 行文本编辑器缓冲区中的最后追加指定内容，例：leAppendLine $result "abc"
	"leAppendLine":  70031, // 等同于leAppend
	"leSet":         70033, // 设定行文本编辑器缓冲区中的指定行为指定内容，例：leSet  $result 3 "abc"
	"leSetLine":     70033, // 设定行文本编辑器缓冲区中的指定行为指定内容，例：leSetLine $result 3 "abc"
	"leSetLines":    70037, // 设定行文本编辑器缓冲区中指定范围的多行为指定内容，例：leSetLines $result 3 5 "abc\nbbb"
	"leRemove":      70039, // 删除行文本编辑器缓冲区中的指定行，例：leRemove $result 3
	"leRemoveLine":  70039, // 等同于leRemove
	"leRemoveLines": 70043, // 删除行文本编辑器缓冲区中指定范围的多行，例：leRemoveLines $result 1 3
	"leViewAll":     70045, // 查看行文本编辑器缓冲区中的所有内容，例：leViewAll $textT
	"leView":        70047, // 查看行文本编辑器缓冲区中的指定行，例：leView $lineText 18
	"leViewLine":    70047,
	"leSort":        70049, // 将行文本编辑器缓冲区中的行进行排序，唯一参数表示是否降序排序，例：leSort $result true
	"leEnc":         70051, // 将行文本编辑器缓冲区中的文本转换为UTF-8编码，如果不指定原始编码则默认为GB18030编码，用法：leEnc $result gbk
	"leLineEnd":     70061, // 读取或设置行文本编辑器缓冲区中行末字符（一般是\n或\r\n），不带参数是获取，带参数是设置
	"leSilent":      70071, // 读取或设置行文本编辑器的静默模式（布尔值），不带参数是获取，带参数是设置
	"leFind":        70081, // 在编辑缓冲区查找包含某字符串（可以是正则表达式）的行
	"leReplace":     70083, // 在编辑缓冲区查找包含某字符串（可以是正则表达式）的行并替换相关内容

	"leSSHInfo": 70091, // 获取当前行文本编辑器使用的SSH连接的信息

	"leRun": 70098, // 将当前行文本编辑器中保存的文本作为谢语言代码执行，结果参数不可省略，如有第二个参数则为要传入的inputG，第三个参数开始为传入的argsG

	// server related 服务器相关
	"getMimeType": 80001, // 根据文件名获取MIME类型，文件名可以包含路径
	"getMIMEType": 80001,

	// zip related 压缩相关
	"archiveFilesToZip":   90101, // 添加多个文件到一个新建的zip文件，第一个参数为zip文件名，后缀必须是.zip，可选参数-overwrite（是否覆盖已有文件），-makeDirs（是否根据需要新建目录），其他参数看做是需要添加的文件或目录，目录将递归加入zip文件，如果参数为一个列表，将看作一个文件名列表，其中的文件都将加入
	"extractFilesFromZip": 90111, // 添加文件到zip文件

	// web GUI related 网页界面相关
	"initWebGUIW":   100001, // 初始化Web图形界面编程环境（Windows下IE11版本），如果没有外嵌式浏览器xiewbr，则将其下载到xie语言目录下
	"initWebGuiW":   100001,
	"updateWebGuiW": 100003, // 强制刷新Web图形界面编程环境（Windows下IE11版本），会将最新的外嵌式浏览器xiewbr下载到xie语言目录下

	"initWebGUIC": 100011, // 初始化Web图形界面编程环境（Windows下CEF版本），如果没有外嵌式浏览器xiecbr及相关库文件，则将其下载到xie语言目录下
	"initWebGuiC": 100011,

	// ssh/sftp/ftp related
	"sshUpload": 200011, // 通过ssh上传一个文件，用法：sshUpload 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -path=文件路径 -remotePath=远端文件路径

	// GUI related 图形界面相关
	// "guiInit": 210011,

	// "guiNewApp": 210013,

	// "guiSetFont": 210021,

	// "guiNewWindow": 210031,

	// "guiNewLoop": 210032,

	// "guiCloseWindow": 210033,

	// // "guiSetContent": 210033,

	// "guiRunLoop": 210037,

	// "guiLoopRet": 210038,

	// "guiNewFunc": 210041,

	// "guiLayout": 210051,

	// "guiSetVar": 210101,

	// "guiSetVarByRef": 210102,

	// "guiNewVar": 210103,

	// "guiGetVar": 210105,

	// "guiGetVarByRef": 210107,

	// "guiNewLabel": 211001,

	// "guiStaticLabel": 211002,

	// "guiLabel": 211003,

	// // "guiNewButton": 211011,

	// "guiButton":       211013,
	// "guiStaticButton": 211014,

	// "guiInput": 211015,

	// "guiRow": 211101,

	// "guiColumn": 211103,

	// "guiSpacing": 211105,
	// "guiGap":     211105,

	// "guiNewVBox": 211101,

	// end of commands/instructions 指令集末尾
}

type VarRef struct {
	Ref   int // -99 - invalid, -15 - ref, -12 - unref, -10 - quickEval, -9 - eval, -8 - pop, -7 - peek, -6 - push, -5 - tmp, -4 - pln, -3 - var(string), -2 - drop, -1 - debug, > 0 normal vars
	Value interface{}
}

func GetVarRefFromArray(aryA []VarRef, idxT int) VarRef {
	if idxT < 0 || idxT >= len(aryA) {
		return VarRef{Ref: -10, Value: nil}
	}

	return aryA[idxT]
}

type Instr struct {
	Code     int
	ParamLen int
	Params   []VarRef
	Line     string
	// Param1Ref   int
	// Param1Value interface{}
	// Param2Ref   int
	// Param2Value interface{}
}

func (v Instr) ParamsToStrs(fromA int) []string {

	lenT := len(v.Params)

	sl := make([]string, 0, lenT)

	for i := fromA; i < lenT; i++ {
		sl = append(sl, tk.ToStr(v.Params[i].Value))
	}

	return sl
}

func (v Instr) ParamsToList(fromA int) []interface{} {

	lenT := len(v.Params)

	sl := make([]interface{}, 0, lenT)

	for i := fromA; i < lenT; i++ {
		sl = append(sl, v.Params[i].Value)
	}

	return sl
}

// type Regs struct {
// 	IntsM   [5]int
// 	FloatsM [5]float64
// 	CondsM  [5]bool
// 	StrsM   [5]string
// 	AnysM   [5]interface{}
// }

type FuncContext struct {
	// VarsM          map[int]interface{}
	VarsLocalMapM  map[int]int
	VarsM          *[]interface{}
	ReturnPointerM int
	// RegsM          *Regs

	DeferStackM *tk.SimpleStack

	Layer int

	// StackM []interface{}
}

type XieVM struct {
	SourceM        []string
	CodeListM      []string
	InstrListM     []Instr
	CodeSourceMapM map[int]int

	LabelsM      map[int]int
	VarIndexMapM map[string]int
	VarNameMapM  map[int]string

	CodePointerM int

	StackM        []interface{}
	StackPointerM int

	// StackM *linkedliststack.Stack

	FuncStackM        []FuncContext
	FuncStackPointerM int

	// VarsM map[int]interface{}
	// VarsLocalMapM map[int]int
	// VarsM []interface{}

	// RegsM Regs

	FuncContextM FuncContext

	// CurrentRegsM *Regs
	// // CurrentVarsM *(map[int]interface{})
	// CurrentVarsM *([]interface{})

	CurrentFuncContextM *FuncContext

	TmpM interface{} // 预制全局变量$tmp，一般用于临时存储

	SharedMapM *tk.SyncMap // 预设全局共享映射，线程安全，可用于存储各个虚拟机需共享的内容

	ErrorHandlerM int

	VerboseM bool

	VerbosePlusM bool

	GuiM map[string]interface{}
	// GuiStrVarsM   []string
	// GuiIntVarsM   []int
	// GuiFloatVarsM []int
}

type XieObject interface {
	TypeName() string

	Init(argsA ...interface{}) interface{}

	SetValue(argsA ...interface{}) interface{}

	GetValue(argsA ...interface{}) interface{}

	Call(argsA ...interface{}) interface{}

	SetMember(argsA ...interface{}) interface{}

	GetMember(argsA ...interface{}) interface{}
}

// type XieObjectImpl struct {
// 	Members map[string]XieObject
// 	Methods map[string]*XieObject
// }

// var _ XieObject = XieObjectImpl{}

type XieString struct {
	Value string

	Members map[string]interface{}
}

func (v XieString) TypeName() string {
	return "string"
}

func (v XieString) String() string {
	return v.Value
}

func (p *XieString) Init(argsA ...interface{}) interface{} {
	if len(argsA) > 0 {
		p.Value = tk.ToStr(argsA[0])
	} else {
		p.Value = ""
	}

	return nil
}

func (p *XieString) SetValue(argsA ...interface{}) interface{} {
	if len(argsA) > 0 {
		p.Value = tk.ToStr(argsA[0])
	} else {
		p.Value = ""
	}

	return nil
}

func (v XieString) GetValue(argsA ...interface{}) interface{} {
	return v.Value
}

func (p *XieString) SetMember(argsA ...interface{}) interface{} {
	// if len(argsA) < 2 {
	// 	return fmt.Errorf("参数个数不够")
	// }

	// return nil
	return fmt.Errorf("不支持此方法")
}

func (p *XieString) GetMember(argsA ...interface{}) interface{} {
	// if len(argsA) < 2 {
	// 	return fmt.Errorf("参数个数不够")
	// }

	// return nil
	return fmt.Errorf("不支持此方法")
}

func (p *XieString) Call(argsA ...interface{}) interface{} {
	if len(argsA) < 1 {
		return fmt.Errorf("参数个数不够")
	}

	methodNameT := tk.ToStr(argsA[0])

	switch methodNameT {
	case "toStr":
		return p.Value
	case "len":
		return len(p.Value)
	case "toRuneArray":
		return []rune(p.Value)
	case "runeLen":
		return len([]rune(p.Value))
	case "toByteArray":
		return []byte(p.Value)
	case "trim":
		p.Value = strings.TrimSpace(p.Value)
		return nil
	case "trimSet":
		if len(argsA) < 2 {
			return fmt.Errorf("参数个数不够")
		}

		p.Value = strings.Trim(p.Value, tk.ToStr(argsA[1]))
		return nil
	case "add":
		if len(argsA) < 2 {
			return fmt.Errorf("参数个数不够")
		}

		p.Value = p.Value + tk.ToStr(argsA[1])

		return nil
	default:
		return fmt.Errorf("未知方法")
	}

	return nil
}

// type XieHandler struct {
// 	Value string

// 	Members map[string]interface{}
// }

// func (v XieString) TypeName() string {
// 	return "string"
// }

// func (v XieString) String() string {
// 	return v.Value
// }

// func (p *XieString) Init(argsA ...interface{}) interface{} {
// 	if len(argsA) > 0 {
// 		p.Value = tk.ToStr(argsA[0])
// 	} else {
// 		p.Value = ""
// 	}

// 	return nil
// }

// func (p *XieString) SetValue(argsA ...interface{}) interface{} {
// 	if len(argsA) > 0 {
// 		p.Value = tk.ToStr(argsA[0])
// 	} else {
// 		p.Value = ""
// 	}

// 	return nil
// }

// func (v XieString) GetValue(argsA ...interface{}) interface{} {
// 	return v.Value
// }

// func (p *XieString) SetMember(argsA ...interface{}) interface{} {
// 	// if len(argsA) < 2 {
// 	// 	return fmt.Errorf("参数个数不够")
// 	// }

// 	// return nil
// 	return fmt.Errorf("不支持此方法")
// }

// func (p *XieString) GetMember(argsA ...interface{}) interface{} {
// 	// if len(argsA) < 2 {
// 	// 	return fmt.Errorf("参数个数不够")
// 	// }

// 	// return nil
// 	return fmt.Errorf("不支持此方法")
// }

// func (p *XieString) Call(argsA ...interface{}) interface{} {
// 	if len(argsA) < 1 {
// 		return fmt.Errorf("参数个数不够")
// 	}

// 	methodNameT := tk.ToStr(argsA[0])

// 	switch methodNameT {
// 	case "toStr":
// 		return p.Value
// 	case "len":
// 		return len(p.Value)
// 	case "toRuneArray":
// 		return []rune(p.Value)
// 	case "runeLen":
// 		return len([]rune(p.Value))
// 	case "toByteArray":
// 		return []byte(p.Value)
// 	case "trim":
// 		p.Value = strings.TrimSpace(p.Value)
// 		return nil
// 	case "trimSet":
// 		if len(argsA) < 2 {
// 			return fmt.Errorf("参数个数不够")
// 		}

// 		p.Value = strings.Trim(p.Value, tk.ToStr(argsA[1]))
// 		return nil
// 	case "add":
// 		if len(argsA) < 2 {
// 			return fmt.Errorf("参数个数不够")
// 		}

// 		p.Value = p.Value + tk.ToStr(argsA[1])

// 		return nil
// 	default:
// 		return fmt.Errorf("未知方法")
// 	}

// 	return nil
// }

type XieAny struct {
	Value interface{}

	Members map[string]interface{}
}

func (v XieAny) TypeName() string {
	return "any"
}

func (v XieAny) String() string {
	return tk.ToJSONX(v)
}

func (p *XieAny) Init(argsA ...interface{}) interface{} {
	if len(argsA) > 0 {
		p.Value = argsA[0]
	} else {
		p.Value = nil
	}

	p.Members = make(map[string]interface{})

	return nil
}

func (p *XieAny) SetValue(argsA ...interface{}) interface{} {
	if len(argsA) > 0 {
		p.Value = argsA[0]
	} else {
		p.Value = ""
	}

	return nil
}

func (v XieAny) GetValue(argsA ...interface{}) interface{} {
	return v.Value
}

func (p *XieAny) SetMember(argsA ...interface{}) interface{} {
	if len(argsA) < 2 {
		return fmt.Errorf("参数个数不够")
	}

	p.Members[tk.ToStr(argsA[0])] = argsA[1]

	return nil
}

func (p *XieAny) GetMember(argsA ...interface{}) interface{} {
	if len(argsA) < 1 {
		return fmt.Errorf("参数个数不够")
	}

	nv, ok := p.Members[tk.ToStr(argsA[0])]

	if !ok {
		return Undefined
	}

	return nv
}

func (p *XieAny) Call(argsA ...interface{}) interface{} {
	if len(argsA) < 1 {
		return fmt.Errorf("参数个数不够")
	}

	methodNameT := tk.ToStr(argsA[0])

	switch methodNameT {
	case "toStr":
		return p.String()
	default:
		return fmt.Errorf("未知方法")
	}

	return nil
}

func fnASRSE(fn func(string) (string, error)) func(args ...interface{}) interface{} {
	return func(args ...interface{}) interface{} {
		if len(args) != 1 {
			return tk.Errf("not enough parameters")
		}

		s := tk.ToStr(args[0])

		strT, errT := fn(s)

		if errT != nil {
			return errT
		}

		return strT
	}
}

func NewXie(sharedMapA *tk.SyncMap, globalsA ...map[string]interface{}) *XieVM {
	vmT := &XieVM{}

	vmT.InitVM(sharedMapA, globalsA...)

	return vmT
}

func (p *XieVM) InitVM(sharedMapA *tk.SyncMap, globalsA ...map[string]interface{}) {
	p.ErrorHandlerM = -1

	p.StackM = make([]interface{}, 0, 10)
	p.StackPointerM = 0

	p.FuncStackM = make([]FuncContext, 0, 10)

	p.VarIndexMapM = make(map[string]int, 100)
	p.VarNameMapM = make(map[int]string, 100)

	// p.VarsM = make(map[int]interface{}, 100)
	// p.VarsM = make([]interface{}, 0, 100)
	// p.VarsLocalMapM = make(map[int]string, 100)

	// p.CurrentFuncContextM.RegsM = &(p.RegsM)
	// p.CurrentVarsM = &(p.VarsM)

	// p.FuncContextM = FuncContext{VarsM: make([]interface{}, 0, 10), VarsLocalMapM: make(map[int]int, 10), ReturnPointerM: -1}
	p.FuncContextM = FuncContext{VarsM: &([]interface{}{}), VarsLocalMapM: make(map[int]int, 10), ReturnPointerM: -1}

	p.CurrentFuncContextM = &(p.FuncContextM)

	p.CurrentFuncContextM.DeferStackM = tk.NewSimpleStack()

	p.SetVar("backQuoteG", "`")
	p.SetVar("undefined", Undefined)
	p.SetVar("newLineG", "\n")
	p.SetVar("tmp", "")

	if len(globalsA) > 0 {
		globalsT := globalsA[0]

		for k, v := range globalsT {
			p.SetVar(k, v)
		}
	}

	p.SourceM = make([]string, 0, 100)

	p.CodeListM = make([]string, 0, 100)
	p.InstrListM = make([]Instr, 0, 100)

	p.LabelsM = make(map[int]int, 100)

	p.CodeSourceMapM = make(map[int]int, 100)

	if sharedMapA == nil {
		p.SharedMapM = tk.NewSyncMap()
	} else {
		p.SharedMapM = sharedMapA
	}

	p.GuiM = make(map[string]interface{}, 10)
	// p.GuiStrVarsM = make([]string, 0, 10)
	// p.GuiIntVarsM = make([]int, 0, 10)
	// p.GuiFloatVarsM = make([]int, 0, 10)

	if tk.IfSwitchExistsWhole(os.Args, "-vv") {
		p.VerbosePlusM = true
	}

	if tk.IfSwitchExistsWhole(os.Args, "-verbose") {
		p.VerboseM = true
	}

}

func (p *XieVM) ParseVar(strA string) VarRef {
	s1T := strings.TrimSpace(strA)

	if strings.HasPrefix(s1T, "`") && strings.HasSuffix(s1T, "`") {
		s1T = s1T[1 : len(s1T)-1]

		return VarRef{-3, s1T} // value(string)
	} else if strings.HasPrefix(s1T, `"`) && strings.HasSuffix(s1T, `"`) {
		tmps, errT := strconv.Unquote(s1T)

		if errT != nil {
			return VarRef{-3, s1T}
		}

		return VarRef{-3, tmps} // value(string)
	} else {
		if strings.HasPrefix(s1T, "$") {
			if s1T == "$drop" || s1T == "$丢弃" {
				return VarRef{-2, nil}
			} else if s1T == "$debug" || s1T == "$调试" {
				return VarRef{-1, nil}
			} else if s1T == "$pln" || s1T == "$行输出" {
				return VarRef{-4, nil}
			} else if s1T == "$pop" || s1T == "$出栈" {
				return VarRef{-8, nil}
			} else if s1T == "$peek" || s1T == "$看栈" {
				return VarRef{-7, nil}
			} else if s1T == "$push" || s1T == "$入栈" {
				return VarRef{-6, nil}
			} else if s1T == "$tmp" || s1T == "$临时变量" {
				return VarRef{-5, nil}
			} else {
				vNameT := s1T[1:]

				// if strings.HasPrefix(vNameT, "$") {
				// 	vNameT = vNameT[1:]

				// 	varIndexT, ok := p.VarIndexMapM[vNameT]

				// 	if !ok {
				// 		varIndexT = len(p.VarIndexMapM) + 10000 + 1
				// 		p.VarIndexMapM[vNameT] = varIndexT
				// 		p.VarNameMapM[varIndexT] = vNameT
				// 	}

				// 	return VarRef{varIndexT, nil}
				// }

				varIndexT, ok := p.VarIndexMapM[vNameT]

				if !ok {
					varIndexT = len(p.VarIndexMapM)
					p.VarIndexMapM[vNameT] = varIndexT
					p.VarNameMapM[varIndexT] = vNameT
				}

				return VarRef{varIndexT, nil}
			}
		} else if strings.HasPrefix(s1T, "*") {
			vNameT := s1T[1:]

			varIndexT, ok := p.VarIndexMapM[vNameT]

			if !ok {
				varIndexT = len(p.VarIndexMapM)
				p.VarIndexMapM[vNameT] = varIndexT
				p.VarNameMapM[varIndexT] = vNameT
			}

			return VarRef{-12, varIndexT}
		} else if strings.HasPrefix(s1T, "&") {
			vNameT := s1T[1:]

			varIndexT, ok := p.VarIndexMapM[vNameT]

			if !ok {
				varIndexT = len(p.VarIndexMapM)
				p.VarIndexMapM[vNameT] = varIndexT
				p.VarNameMapM[varIndexT] = vNameT
			}

			return VarRef{-15, varIndexT}
		} else if strings.HasPrefix(s1T, ":") { // labels
			vNameT := s1T[1:]
			varIndexT, ok := p.VarIndexMapM[vNameT]

			if !ok {
				return VarRef{-3, s1T}
			}

			return VarRef{-3, p.LabelsM[varIndexT]}
		} else if strings.HasPrefix(s1T, "#") { // values
			if len(s1T) < 2 {
				return VarRef{-3, s1T}
			}

			// remainsT := s1T[2:]

			typeT := s1T[1]

			if typeT == 'i' {
				c1T, errT := tk.StrToIntQuick(s1T[2:])

				if errT != nil {
					return VarRef{-3, s1T}
				}

				return VarRef{-3, c1T}
			} else if typeT == 'f' {
				c1T, errT := tk.StrToFloat64E(s1T[2:])

				if errT != nil {
					return VarRef{-3, s1T}
				}

				return VarRef{-3, c1T}
			} else if typeT == 'b' {
				return VarRef{-3, tk.ToBool(s1T[2:])}
			} else if typeT == 'y' {
				return VarRef{-3, tk.ToByte(s1T[2:])}
			} else if typeT == 'r' {
				return VarRef{-3, tk.ToRune(s1T[2:])}
			} else if typeT == 's' {
				s1DT := s1T[2:]

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				}

				return VarRef{-3, tk.ToStr(s1DT)}
			} else if typeT == 'e' {
				s1DT := s1T[2:]

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				}

				return VarRef{-3, fmt.Errorf("%v", s1DT)}
			} else if typeT == 't' {
				s1DT := s1T[2:]

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				}

				tmps := strings.TrimSpace(s1DT)

				if tmps == "" || tmps == "now" {
					return VarRef{-3, time.Now()}
				}

				rsT := tk.ToTime(tmps)

				if tk.IsError(rsT) {
					return VarRef{-3, s1T}
				}

				return VarRef{-3, rsT}
			} else if typeT == 'L' { // list
				var listT []interface{}

				s1DT := s1T[2:] // tk.UrlDecode(s1T[2:])

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				}

				// tk.Plv(s1T[2:])
				// tk.Plv(s1DT)

				errT := json.Unmarshal([]byte(s1DT), &listT)
				// tk.Plv(errT)
				if errT != nil {
					return VarRef{-3, s1T}
				}

				// tk.Plv(listT)
				return VarRef{-3, listT}
			} else if typeT == 'Y' { // byteList
				var listT []byte

				s1DT := s1T[2:] // tk.UrlDecode(s1T[2:])

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				}

				// tk.Plv(s1T[2:])
				// tk.Plv(s1DT)

				errT := json.Unmarshal([]byte(s1DT), &listT)
				// tk.Plv(errT)
				if errT != nil {
					return VarRef{-3, s1T}
				}

				// tk.Plv(listT)
				return VarRef{-3, listT}
			} else if typeT == 'R' { // runeList
				var listT []rune

				s1DT := s1T[2:] // tk.UrlDecode(s1T[2:])

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				}

				// tk.Plv(s1T[2:])
				// tk.Plv(s1DT)

				errT := json.Unmarshal([]byte(s1DT), &listT)
				// tk.Plv(errT)
				if errT != nil {
					return VarRef{-3, s1T}
				}

				// tk.Plv(listT)
				return VarRef{-3, listT}
			} else if typeT == 'S' { // strList/stringList
				var listT []string

				s1DT := s1T[2:] // tk.UrlDecode(s1T[2:])

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				}

				// tk.Plv(s1T[2:])
				// tk.Plv(s1DT)

				errT := json.Unmarshal([]byte(s1DT), &listT)
				// tk.Plv(errT)
				if errT != nil {
					return VarRef{-3, s1T}
				}

				// tk.Plv(listT)
				return VarRef{-3, listT}
			} else if typeT == 'M' { // map
				var mapT map[string]interface{}

				s1DT := s1T[2:] // tk.UrlDecode(s1T[2:])

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				}

				// tk.Plv(s1T[2:])
				// tk.Plv(s1DT)

				errT := json.Unmarshal([]byte(s1DT), &mapT)
				// tk.Plv(errT)
				if errT != nil {
					return VarRef{-3, s1T}
				}

				// tk.Plv(listT)
				return VarRef{-3, mapT}
			}

			return VarRef{-3, s1T}
			// } else if strings.HasPrefix(s1T, "@") { // regs
			// 	if len(s1T) < 2 {
			// 		return VarRef{-3, s1T}
			// 	}

			// 	typeT := s1T[1]

			// 	if typeT == 'i' {
			// 		c1T, errT := tk.StrToIntQuick(s1T[2:])

			// 		if errT != nil {
			// 			return VarRef{-3, s1T}
			// 		}

			// 		return VarRef{-220 - c1T, nil}
			// 	} else if typeT == 'f' {
			// 		c1T, errT := tk.StrToIntQuick(s1T[2:])

			// 		if errT != nil {
			// 			return VarRef{-3, s1T}
			// 		}

			// 		return VarRef{-230 - c1T, nil}
			// 	} else if typeT == 'b' {
			// 		c1T, errT := tk.StrToIntQuick(s1T[2:])

			// 		if errT != nil {
			// 			return VarRef{-3, s1T}
			// 		}

			// 		return VarRef{-210 - c1T, nil}
			// 	} else if typeT == 's' {
			// 		c1T, errT := tk.StrToIntQuick(s1T[2:])

			// 		if errT != nil {
			// 			return VarRef{-3, s1T}
			// 		}

			// 		return VarRef{-240 - c1T, nil}
			// 	} else if typeT == 'a' {
			// 		c1T, errT := tk.StrToIntQuick(s1T[2:])

			// 		if errT != nil {
			// 			return VarRef{-3, s1T}
			// 		}

			// 		return VarRef{-250 - c1T, nil}
			// 	}

			// 	return VarRef{-3, s1T}
		} else if strings.HasPrefix(s1T, "?") { // eval
			if len(s1T) < 2 {
				return VarRef{-3, s1T}
			}

			s1T = strings.TrimSpace(s1T[1:])

			if strings.HasPrefix(s1T, "`") && strings.HasSuffix(s1T, "`") {
				s1T = s1T[1 : len(s1T)-1]

				return VarRef{-9, s1T} // eval value
			} else if strings.HasPrefix(s1T, `"`) && strings.HasSuffix(s1T, `"`) {
				tmps, errT := strconv.Unquote(s1T)

				if errT != nil {
					return VarRef{-9, s1T}
				}

				return VarRef{-9, tmps}
			}

			return VarRef{-9, s1T}
			// } else if strings.HasPrefix(s1T, "-") { // switch like -format=json, -index=%v
			// 	groupsT := tk.RegFindFirstGroupsX(s1T, `^(-\S+=)(\$.*)$`)

			// 	if groupsT != nil {
			// 		tk.Plv(groupsT)
			// 		tk.Plv(p.ParseVar(groupsT[2]))
			// 		tmps := tk.ToStr(p.GetVarValue(p.ParseVar(groupsT[2])))

			// 		tmps = "`" + strings.ReplaceAll(tmps, "`", "~~~") + "`"

			// 		return VarRef{-3, groupsT[1] + tmps}
			// 	}

			// 	return VarRef{-3, s1T}
		} else if strings.HasPrefix(s1T, "@") { // quickEval
			if len(s1T) < 2 {
				return VarRef{-3, s1T}
			}

			s1T = strings.TrimSpace(s1T[1:])

			if strings.HasPrefix(s1T, "`") && strings.HasSuffix(s1T, "`") {
				s1T = s1T[1 : len(s1T)-1]

				return VarRef{-10, s1T} // eval value
			} else if strings.HasPrefix(s1T, `"`) && strings.HasSuffix(s1T, `"`) {
				tmps, errT := strconv.Unquote(s1T)

				if errT != nil {
					return VarRef{-10, s1T}
				}

				return VarRef{-10, tmps}
			}

			return VarRef{-10, s1T}
		} else {
			return VarRef{-3, s1T} // value(string)
		}
	}
}

func isOperator(strA string) (bool, string) {
	return false, ""
}

func evalSingle(exprA []interface{}) (resultR interface{}) {
	// tk.Plvx(exprA)
	resultR = nil

	opT := exprA[1].(string)

	if exprA[0] == nil {
		if opT == "-" {
			switch nv := exprA[2].(type) {
			case int:
				resultR = -nv
				return
			case float64:
				resultR = -nv
				return
			case byte:
				resultR = -nv
				return
			case rune:
				resultR = -nv
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}

		} else if opT == "!" {
			if tk.IsNil(exprA[2]) {
				return true
			}

			switch nv := exprA[2].(type) {
			case UndefinedStruct:
				resultR = true
				return
			case bool:
				resultR = !nv
				return
			default:
				resultR = false
				// resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}

		}

		resultR = fmt.Errorf("未知运算符：%v", opT)
		return
	} else {
		if opT == "+" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv + exprA[2].(int)
				return
			case float64:
				resultR = nv + exprA[2].(float64)
				return
			case byte:
				resultR = nv + exprA[2].(byte)
				return
			case rune:
				resultR = nv + exprA[2].(rune)
				return
			case string:
				resultR = nv + exprA[2].(string)
				return
			case time.Time:
				resultR = nv.Add(time.Duration(time.Millisecond * time.Duration(tk.ToInt(exprA[2]))))
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T （%v）-> %T", exprA[0], exprA[1], exprA[2])
				return
			}
		} else if opT == "-" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv - exprA[2].(int)
				return
			case float64:
				resultR = nv - exprA[2].(float64)
				return
			case byte:
				resultR = nv - exprA[2].(byte)
				return
			case rune:
				resultR = nv - exprA[2].(rune)
				return
			case time.Time:
				t2 := tk.ToTime(exprA[2])
				if tk.IsError(t2) {
					t2 := tk.ToInt(exprA[2], tk.MAX_INT)

					if t2 == tk.MAX_INT {
						resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[2])
						return
					}

					resultR = nv.Add(time.Duration(-t2) * time.Millisecond)
					return
				}

				resultR = tk.ToInt(nv.Sub(t2.(time.Time)) / time.Millisecond)
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "*" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv * exprA[2].(int)
				return
			case float64:
				resultR = nv * exprA[2].(float64)
				return
			case byte:
				resultR = nv * exprA[2].(byte)
				return
			case rune:
				resultR = nv * exprA[2].(rune)
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "/" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv / exprA[2].(int)
				return
			case float64:
				resultR = nv / exprA[2].(float64)
				return
			case byte:
				resultR = nv / exprA[2].(byte)
				return
			case rune:
				resultR = nv / exprA[2].(rune)
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "%" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv % exprA[2].(int)
				return
			case byte:
				resultR = nv % exprA[2].(byte)
				return
			case rune:
				resultR = nv % exprA[2].(rune)
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "<" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv < exprA[2].(int)
				return
			case byte:
				resultR = nv < exprA[2].(byte)
				return
			case rune:
				resultR = nv < exprA[2].(rune)
				return
			case float64:
				resultR = nv < exprA[2].(float64)
				return
			case string:
				resultR = nv < exprA[2].(string)
				return
			case time.Time:
				rsT := tk.ToTime(exprA[2])
				if tk.IsError(rsT) {
					resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
					return
				}
				resultR = nv.Before(rsT.(time.Time))
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "<=" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv <= exprA[2].(int)
				return
			case byte:
				resultR = nv <= exprA[2].(byte)
				return
			case rune:
				resultR = nv <= exprA[2].(rune)
				return
			case float64:
				resultR = nv <= exprA[2].(float64)
				return
			case string:
				resultR = nv <= exprA[2].(string)
				return
			case time.Time:
				rsT := tk.ToTime(exprA[2])
				if tk.IsError(rsT) {
					resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
					return
				}
				resultR = nv.Before(rsT.(time.Time)) || nv.Equal(rsT.(time.Time))
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == ">" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv > exprA[2].(int)
				return
			case byte:
				resultR = nv > exprA[2].(byte)
				return
			case rune:
				resultR = nv > exprA[2].(rune)
				return
			case float64:
				resultR = nv > exprA[2].(float64)
				return
			case string:
				resultR = nv > exprA[2].(string)
				return
			case time.Time:
				rsT := tk.ToTime(exprA[2])
				if tk.IsError(rsT) {
					resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
					return
				}
				resultR = nv.After(rsT.(time.Time))
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == ">=" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv >= exprA[2].(int)
				return
			case byte:
				resultR = nv >= exprA[2].(byte)
				return
			case rune:
				resultR = nv >= exprA[2].(rune)
				return
			case float64:
				resultR = nv >= exprA[2].(float64)
				return
			case string:
				resultR = nv >= exprA[2].(string)
				return
			case time.Time:
				rsT := tk.ToTime(exprA[2])
				if tk.IsError(rsT) {
					resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
					return
				}
				resultR = nv.After(rsT.(time.Time)) || nv.Equal(rsT.(time.Time))
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "==" {
			switch nv := exprA[0].(type) {
			case UndefinedStruct:
				_, nvvok := exprA[2].(UndefinedStruct)

				resultR = nvvok
				return
			case bool:
				resultR = nv == exprA[2].(bool)
				return
			case int:
				resultR = nv == exprA[2].(int)
				return
			case byte:
				resultR = nv == exprA[2].(byte)
				return
			case rune:
				resultR = nv == exprA[2].(rune)
				return
			case float64:
				resultR = nv == exprA[2].(float64)
				return
			case string:
				nnv, nnvok := exprA[2].(string)

				if !nnvok {
					resultR = false
					return
				}
				resultR = nv == nnv
				return
			case time.Time:
				rsT := tk.ToTime(exprA[2])
				if tk.IsError(rsT) {
					resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
					return
				}
				resultR = nv.Equal(rsT.(time.Time))
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "!=" || opT == "<>" {
			switch nv := exprA[0].(type) {
			case UndefinedStruct:
				_, nvvok := exprA[2].(UndefinedStruct)

				resultR = !nvvok
				return
			case bool:
				resultR = nv != exprA[2].(bool)
				return
			case int:
				resultR = nv != exprA[2].(int)
				return
			case byte:
				resultR = nv != exprA[2].(byte)
				return
			case rune:
				resultR = nv != exprA[2].(rune)
				return
			case float64:
				resultR = nv != exprA[2].(float64)
				return
			case string:
				resultR = nv != exprA[2].(string)
				return
			case time.Time:
				rsT := tk.ToTime(exprA[2])
				if tk.IsError(rsT) {
					resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
					return
				}
				resultR = !nv.Equal(rsT.(time.Time))
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "&&" {
			switch nv := exprA[0].(type) {
			case bool:
				resultR = nv && exprA[2].(bool)
				return
			default:
				resultR = fmt.Errorf("无法处理的类型：%T", exprA[0])
				return
			}
		} else if opT == "||" {
			switch nv := exprA[0].(type) {
			case bool:
				resultR = nv || exprA[2].(bool)
				return
			default:
				resultR = fmt.Errorf("无法处理的类型：%T", exprA[0])
				return
			}
		} else if opT == "&" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv & exprA[2].(int)
				return
			case byte:
				// tk.Pl("%v -- %v", nv, exprA[2].(byte))
				resultR = nv & exprA[2].(byte)
				return
			case rune:
				resultR = nv & exprA[2].(rune)
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "|" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv | exprA[2].(int)
				return
			case byte:
				resultR = nv | exprA[2].(byte)
				return
			case rune:
				resultR = nv | exprA[2].(rune)
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "^" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv ^ exprA[2].(int)
				return
			case byte:
				resultR = nv ^ exprA[2].(byte)
				return
			case rune:
				resultR = nv ^ exprA[2].(rune)
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == ">>" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv >> tk.ToInt(exprA[2])
				return
			case byte:
				resultR = nv >> tk.ToInt(exprA[2])
				return
			case rune:
				resultR = nv >> tk.ToInt(exprA[2])
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		} else if opT == "<<" {
			switch nv := exprA[0].(type) {
			case int:
				resultR = nv << tk.ToInt(exprA[2])
				return
			case byte:
				resultR = nv << tk.ToInt(exprA[2])
				return
			case rune:
				resultR = nv << tk.ToInt(exprA[2])
				return
			default:
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
				return
			}
		}

		resultR = fmt.Errorf("未知运算符：%v", opT)
		return
	}

	return
}

type ExprElement struct {
	// 0: value, 1: operator, 6: (, 7: ), 9: end
	Type     int
	Priority int
	Value    string
}

var OperatorPriorityMap map[string]int = map[string]int{
	"||": 20,
	"&&": 25,

	"|": 30,
	"^": 33,
	"&": 35,

	"==": 40,
	"!=": 40,
	"<>": 40,

	">":  50,
	"<":  50,
	">=": 50,
	"<=": 50,

	">>": 60,
	"<<": 60,

	"+": 70,
	"-": 70,

	"*": 80,
	"/": 80,
	"%": 80,

	"1!": 99,
	"++": 99,
	"--": 99,
	"1+": 99,
	"1-": 99,
	"1&": 99,
	"1*": 99,
	"1^": 99,
}

func (p *XieVM) SplitExpr(strA string) ([]ExprElement, error) {
	runeT := []rune(strings.TrimSpace(strA))

	elementsT := make([]ExprElement, 0)

	// 0: start, 1: operator, 2: value, 3: value in quote such as "abc", 4: wait slash in quote such as "ab\n", 5: blank after value
	stateT := 0

	opT := ""
	valueT := ""

	for _, v := range runeT {
		// tk.Pl("state: %v, v: %v(%v), op: %v, value: %v, list: %v", stateT, v, string(v), opT, valueT, elementsT)
		if stateT == 0 { // 开始或非值后空白（运算符后）
			switch v {
			case ' ':
				break
			case '"':
				stateT = 3
				valueT = `"`
				break
			case '+', '-', '*', '!', '&', '^':
				stateT = 1
				opT = "1" + string(v)
				break
			case '(':
				stateT = 0
				elementsT = append(elementsT, ExprElement{Type: 6, Priority: 0, Value: "("})
				break
			case ')':
				return nil, fmt.Errorf("无法匹配的括号")
				// stateT = 0
				// elementsT = append(elementsT, ExprElement{Type: 7, Priority: 0, Value: ")"})
				// break
			default:
				stateT = 2
				valueT = string(v)
				break
			}

			continue
		} else if stateT == 5 { // 值后的空白
			switch v {
			case ' ':
				break
			case '"':
				stateT = 3
				valueT = `"`
				break
			case '+', '-', '*', '/', '%', '!', '&', '|', '=', '<', '>', '^':
				stateT = 1
				opT = string(v)
				break
			case '(':
				stateT = 0
				elementsT = append(elementsT, ExprElement{Type: 6, Priority: 0, Value: "("})
				break
			case ')':
				stateT = 5
				elementsT = append(elementsT, ExprElement{Type: 7, Priority: 0, Value: ")"})
				break
			default:
				stateT = 2
				valueT = string(v)
				break
			}

			continue
		} else if stateT == 1 { // 进入运算符阶段
			switch v {
			case ' ':
				stateT = 0
				opT = strings.TrimSpace(opT)
				if len(opT) > 0 {
					elementsT = append(elementsT, ExprElement{Type: 1, Priority: OperatorPriorityMap[opT], Value: opT})
				}
				opT = ""
				break
			case '+', '-', '*', '/', '%', '!', '&', '|', '=', '<', '>', '^':
				opT += string(v)
				break
			case '(':
				stateT = 0
				opT = strings.TrimSpace(opT)
				if len(opT) > 0 {
					elementsT = append(elementsT, ExprElement{Type: 1, Priority: OperatorPriorityMap[opT], Value: opT})
				}
				opT = ""
				elementsT = append(elementsT, ExprElement{Type: 6, Priority: 0, Value: "("})
				break
			case ')':
				stateT = 5
				opT = strings.TrimSpace(opT)
				if len(opT) > 0 {
					elementsT = append(elementsT, ExprElement{Type: 1, Priority: OperatorPriorityMap[opT], Value: opT})
				}
				opT = ""
				elementsT = append(elementsT, ExprElement{Type: 7, Priority: 0, Value: ")"})
				break
			default:
				stateT = 2
				opT = strings.TrimSpace(opT)
				if len(opT) > 0 {
					elementsT = append(elementsT, ExprElement{Type: 1, Priority: OperatorPriorityMap[opT], Value: opT})
				}
				opT = ""
				valueT = string(v)
				break
			}

			continue
		} else if stateT == 2 { // 进入值阶段
			switch v {
			case '+', '-', '*', '/', '%', '!', '&', '|', '=', '<', '>', '^':
				stateT = 1
				opT = string(v)

				valueT = strings.TrimSpace(valueT)
				if len(valueT) > 0 {
					elementsT = append(elementsT, ExprElement{Type: 0, Priority: 0, Value: valueT})
				}
				valueT = ""
				break
			case '(':
				stateT = 0
				valueT = strings.TrimSpace(valueT)
				if len(valueT) > 0 {
					elementsT = append(elementsT, ExprElement{Type: 0, Priority: 0, Value: valueT})
				}
				valueT = ""
				elementsT = append(elementsT, ExprElement{Type: 6, Priority: 0, Value: "("})
				break
			case ')':
				stateT = 5
				valueT = strings.TrimSpace(valueT)
				if len(valueT) > 0 {
					elementsT = append(elementsT, ExprElement{Type: 0, Priority: 0, Value: valueT})
				}
				valueT = ""
				elementsT = append(elementsT, ExprElement{Type: 7, Priority: 0, Value: ")"})
				break
			default:
				valueT += string(v)
				break
			}

			continue
		} else if stateT == 3 { // 值中的双引号内
			switch v {
			case '"':
				valueT += `"`
				elementsT = append(elementsT, ExprElement{Type: 0, Priority: 0, Value: valueT})
				stateT = 5
				break
			case '\\':
				valueT += string(v)
				stateT = 4
				break
			default:
				valueT += string(v)
				break
			}

			continue
		} else if stateT == 4 { //  值中的双引号内的转义字符
			switch v {
			default:
				// tmps, errT := strconv.Unquote(`"` + "\\" + string(v) + `"`)
				// if errT != nil {
				// 	return nil, errT
				// }
				valueT += string(v) //tmps
				stateT = 3
				break
			}

			continue
		}
	}

	if stateT == 0 {
	} else if stateT == 5 {
	} else if stateT == 1 {
		opT = strings.TrimSpace(opT)
		if len(opT) > 0 {
			elementsT = append(elementsT, ExprElement{Type: 1, Priority: OperatorPriorityMap[opT], Value: opT})
		}
	} else if stateT == 2 {
		valueT = strings.TrimSpace(valueT)
		if len(valueT) > 0 {
			elementsT = append(elementsT, ExprElement{Type: 0, Priority: 0, Value: valueT})
		}
	} else if stateT == 3 || stateT == 4 {
		return nil, fmt.Errorf("表达式格式错误双引号不匹配")
	}

	// tk.Plv(elementsT)

	backElementsT := make([]ExprElement, 0)

	opStackT := tk.NewSimpleStack(len(elementsT) + 1)

	for _, v := range elementsT {
		// tk.Pl("process %v, %v, %v", v, backElementsT, opStackT)
		if v.Type == 0 {
			backElementsT = append(backElementsT, v)
		} else if v.Type == 1 {
			opLast := opStackT.Peek()

			if opLast == nil {
				opStackT.Push(v)
			} else {
				opLastValue := opLast.(ExprElement)
				if v.Priority > opLastValue.Priority {

				} else {
					backElementsT = append(backElementsT, opStackT.Pop().(ExprElement))
				}

				opStackT.Push(v)
			}

			// backElementsT = append(backElementsT, v)
		} else if v.Type == 6 {
			opStackT.Push(v)
			// backElementsT = append(backElementsT, v)
		} else if v.Type == 7 {
			for {
				opLast := opStackT.Pop()

				// tk.Plv(opLast)

				if opLast == nil {
					// tk.Pl("() not match")
					return nil, fmt.Errorf("括号配对错误")
				}

				opLastValue := opLast.(ExprElement)
				if opLastValue.Type == 6 {
					break
				} else {
					backElementsT = append(backElementsT, opLastValue)
				}
			}
		}
	}

	for {
		opLast := opStackT.Pop()

		if opLast == nil {
			break
		}

		opLastValue := opLast.(ExprElement)
		backElementsT = append(backElementsT, opLastValue)
	}

	// tk.Pl("[] %#v", backElementsT)

	return backElementsT, nil
}

func (p *XieVM) QuickEval(strA string) interface{} {

	listT, errT := p.SplitExpr(strA)

	if errT != nil {
		return fmt.Errorf("计算表达式失败：%v", errT)
	}

	valueStackT := tk.NewSimpleStack(len(listT) + 1)

	for _, v := range listT {
		// tk.Pl("item: %v", v)
		if v.Type == 0 {
			v1T := p.ParseVar(v.Value)
			vv1T := p.GetVarValue(v1T)

			valueStackT.Push(vv1T)
		} else if v.Type == 1 {
			switch v.Value {
			case "1-":
				v1 := valueStackT.Pop()

				vr := tk.GetNegativeResult(v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case "1!":
				v1 := valueStackT.Pop()

				vr := tk.GetNegativeResult(v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case "+":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				// tk.Pl("|%#v| |%#v|", v1, v2)

				vr := tk.GetAddResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case "-":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetMinusResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case "*":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetMultiplyResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case "/":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetDivResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case "%":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetModResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case "==":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetEQResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case "!=", "<>":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetNEQResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case ">":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetGTResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case "<":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetLTResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case ">=":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetGETResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			case "<=":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetLETResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("计算表达式失败：%v", vr)
				}

				valueStackT.Push(vr)
			default:
				return fmt.Errorf("未知运算符：%v", v.Value)
			}
		}

		// tk.Pl("vStack: %v", valueStackT)
	}

	vr := valueStackT.Pop()

	return vr
}

func (p *XieVM) EvalExpressionNoGroup(strA string, valuesA *map[string]interface{}) interface{} {
	// strT := strA[1 : len(strA)-1]
	// tk.Pl("EvalExpressionNoGroup: %v", strA)
	// tk.Pl("tmpValues: %v", valuesA)
	if strings.HasPrefix(strA, "?") {
		instrT := p.NewInstr(strA[1:], valuesA)

		if instrT.Code == InstrNameSet["invalidInstr"] {
			return fmt.Errorf("指令分析失败：%v", instrT.Params[0].Value)
		}

		rsT := p.RunLine(0, instrT)

		nsv, ok := rsT.(string)

		if ok {
			if tk.IsErrStr(nsv) {
				return fmt.Errorf("计算失败：%v", tk.GetErrStr(nsv))
			}
		}

		// keyT := "~" + tk.IntToStr(len(valuesA))

		// valuesA[keyT] = p.Pop()

		strA = "$tmp"

	}

	listT := strings.SplitN(strA, " ", -1)

	// lenT := len(listT)

	// opListT := make([][]interface{}, 0, lenT)

	stateT := 0 // 0: initial, 1: first value ready, 2: operator ready, 3: second value ready

	opT := []interface{}{nil, nil, nil}

	// valuesT := make([]interface{})

	for _, v := range listT {
		v = strings.TrimSpace(v)

		if v == "" {
			continue
		}

		// tk.Pl("v: %v", v)

		if tk.InStrings(v, "+", "-", "*", "/", "%", "!", "&&", "||", "==", "!=", "<>", ">", "<", ">=", "<=", "&", "|", "^", ">>", "<<") {
			if stateT == 0 {
				opT[0] = nil
				opT[1] = v
				stateT = 2
			} else if stateT == 1 {
				opT[1] = v
				stateT = 2
			} else if stateT == 2 {
				opT[1] = v
				stateT = 2
			} else {
			}

		} else if strings.HasPrefix(v, "~") {
			if stateT == 0 {
				opT[0] = (*valuesA)[v]
				stateT = 1
			} else if stateT == 1 {
				opT[0] = (*valuesA)[v]
				stateT = 1
			} else if stateT == 2 {
				opT[2] = (*valuesA)[v]
				stateT = 3
			}

		} else {
			vT := p.ParseVar(v)
			vvT := p.GetVarValue(vT)

			if stateT == 0 {
				opT[0] = vvT
				stateT = 1
			} else if stateT == 1 {
				opT[0] = vvT
				stateT = 1
			} else if stateT == 2 {
				opT[2] = vvT
				stateT = 3
			}
		}

		if stateT == 3 {
			// opListT = append(opListT, opT)

			rvT := evalSingle(opT)

			if tk.IsError(rvT) {
				return rvT
			}

			opT[0] = rvT

			stateT = 1
		}

	}

	return opT[0]
}

func (p *XieVM) EvalExpression(strA string) (resultR interface{}) {
	strT := strA
	regexpT := regexp.MustCompile(`\([^\(]*?\)`)

	valuesT := make(map[string]interface{})

	var tmpv interface{}

	for {
		matchT := regexpT.FindStringIndex(strT)

		if matchT == nil {
			tmpv = p.EvalExpressionNoGroup(strT, &valuesT)

			if tk.IsError(tmpv) {
				tk.Pl("表达式计算失败：%v", tmpv)
			}

			break
		} else {
			tmpv = p.EvalExpressionNoGroup(strT[matchT[0]:matchT[1]][1:matchT[1]-matchT[0]-1], &valuesT)

			if tk.IsError(tmpv) {
				tk.Pl("表达式计算失败：%v", tmpv)
			}
		}

		keyT := "~" + tk.IntToStr(len(valuesT))

		valuesT[keyT] = tmpv

		strT = strT[0:matchT[0]] + " " + keyT + " " + strT[matchT[1]:len(strT)]
	}

	resultR = tmpv

	return

	// listT := strings.Split(strA, " ")

	// lenT := len(listT)

	// opListT := make([][]interface{}, lenT)

	// stateT := 0 // 0: initial, 1: first value ready, 2: operator ready, 3: second value ready

	// for i, v := range listT {
	// 	v = strings.TrimSpace(v)

	// 	if v == "" {
	// 		continue
	// 	}

	// 	opT := []interface{}{nil, nil, nil}

	// 	if tk.InStrings(v, "+", "-", "*", "/", "%", "!", "&&", "||", "==", "!=", ">", "<", ">=", "<=") {
	// 		if stateT == 0 {
	// 			opT[0] = nil
	// 			opT[1] = v
	// 			stateT = 2
	// 		}

	// 	} else {
	// 		vT := p.ParseVar(v)
	// 		vvT := p.GetVarValue(vT)

	// 		if stateT == 0 {
	// 			opT[0] = vvT
	// 			stateT = 1
	// 		}
	// 	}

	// 	if stateT == 3 {
	// 		opListT = append(opListT, opT)
	// 	}

	// }

}

func (p *XieVM) GetSwitchVarValue(argsA []string, switchStrA string, defaultA ...string) string {
	vT := tk.GetSwitch(argsA, switchStrA, defaultA...)

	vr := p.ParseVar(vT)

	return tk.ToStr(p.GetVarValue(vr))
}

func (p *XieVM) GetVarValue(vA VarRef) interface{} {
	idxT := vA.Ref

	if idxT == -2 {
		return Undefined
	}

	if idxT == -3 {
		return vA.Value
	}

	if idxT == -5 {
		return p.TmpM
	}

	if idxT == -8 {
		return p.Pop()
	}

	if idxT == -7 {
		return p.Peek()
	}

	if idxT == -1 { // $debug
		return tk.ToJSONX(p, "-indent", "-sort")
	}

	// if idxT < -199 {
	// 	if idxT < -249 {
	// 		return p.CurrentFuncContextM.RegsM.AnysM[(-idxT)-250]
	// 	} else if idxT < -239 {
	// 		return p.CurrentFuncContextM.RegsM.StrsM[(-idxT)-240]
	// 	} else if idxT < -229 {
	// 		return p.CurrentFuncContextM.RegsM.FloatsM[(-idxT)-230]
	// 	} else if idxT < -219 {
	// 		return p.CurrentFuncContextM.RegsM.IntsM[(-idxT)-220]
	// 	} else {
	// 		return p.CurrentFuncContextM.RegsM.CondsM[(-idxT)-210]
	// 	}
	// }

	if idxT == -6 {
		return Undefined
	}

	if idxT == -9 {
		return p.EvalExpression(vA.Value.(string))
	}

	if idxT == -10 {
		return p.QuickEval(vA.Value.(string))
	}

	ifUnrefT := false
	if idxT == -12 { // unref
		ifUnrefT = true

		idxT = tk.ToInt(vA.Value, -99)
	}

	ifRefT := false
	if idxT == -15 { // ref
		ifRefT = true

		idxT = tk.ToInt(vA.Value, -99)
	}

	if idxT < 0 {
		return Undefined
	}

	contextT := p.CurrentFuncContextM

	nv, ok := contextT.VarsLocalMapM[idxT]

	if !ok {

		for {
			if contextT.Layer < 1 {
				break
			}

			if contextT.Layer < 2 {
				contextT = &p.FuncContextM
			} else {
				contextT = &p.FuncStackM[contextT.Layer-2]
			}

			nv, ok = contextT.VarsLocalMapM[idxT]

			if !ok {
				continue
			}

			if ifUnrefT {
				return tk.GetValue((*contextT.VarsM)[nv])
			}

			if ifRefT {
				return &((*contextT.VarsM)[nv])
			}

			return (*contextT.VarsM)[nv]
		}

		return Undefined
	}

	if ifUnrefT {
		return tk.GetValue((*contextT.VarsM)[nv])
	}

	if ifRefT {
		return &((*contextT.VarsM)[nv])
	}

	return (*contextT.VarsM)[nv]

	// vT, ok := (*(p.CurrentVarsM))[idxT]

	// if !ok {
	// 	return Undefined
	// }

	// return vT
}

func (p *XieVM) GetVarValueWithLayer(vA VarRef) (interface{}, int) {
	idxT := vA.Ref

	if idxT == -2 {
		return Undefined, -2
	}

	if idxT == -3 {
		return vA.Value, -1
	}

	if idxT == -5 {
		return p.TmpM, -2
	}

	if idxT == -8 {
		return p.Pop(), -2
	}

	if idxT == -7 {
		return p.Peek(), -2
	}

	if idxT == -6 {
		return Undefined, -2
	}

	if idxT == -1 { // $debug
		return tk.ToJSONX(p, "-indent", "-sort"), -2
	}

	if idxT == -9 {
		return p.EvalExpression(vA.Value.(string)), -2
	}

	if idxT == -10 {
		return p.QuickEval(vA.Value.(string)), -2
	}

	ifUnrefT := false
	if idxT == -12 { // unref
		ifUnrefT = true

		idxT = tk.ToInt(vA.Value, -99)
	}

	ifRefT := false
	if idxT == -15 { // ref
		ifRefT = true

		idxT = tk.ToInt(vA.Value, -99)
	}

	if idxT < 0 {
		return Undefined, -3
	}

	// layerT := len(p.FuncStackM)

	contextT := p.CurrentFuncContextM

	nv, ok := contextT.VarsLocalMapM[idxT]

	if !ok {

		for {
			if contextT.Layer < 1 {
				break
			}

			if contextT.Layer < 2 {
				contextT = &p.FuncContextM
			} else {
				contextT = &p.FuncStackM[contextT.Layer-2]
			}

			// layerT--

			nv, ok = contextT.VarsLocalMapM[idxT]

			if !ok {
				continue
			}

			if ifUnrefT {
				return tk.GetValue((*contextT.VarsM)[nv]), (*contextT).Layer
			}

			if ifRefT {
				return &((*contextT.VarsM)[nv]), (*contextT).Layer
			}

			return (*contextT.VarsM)[nv], (*contextT).Layer
		}

		return Undefined, (*contextT).Layer
	}

	if ifUnrefT {
		return tk.GetValue((*contextT.VarsM)[nv]), (*contextT).Layer
	}

	if ifRefT {
		return &((*contextT.VarsM)[nv]), (*contextT).Layer
	}

	return (*contextT.VarsM)[nv], (*contextT).Layer

	// vT, ok := (*(p.CurrentVarsM))[idxT]

	// if !ok {
	// 	return Undefined
	// }

	// return vT
}

func (p *XieVM) GetVarRef(vA VarRef) *interface{} {
	idxT := vA.Ref

	if idxT == -2 {
		return nil
	}

	if idxT == -3 {
		return nil
	}

	if idxT == -8 {
		return nil
	}

	if idxT == -7 {
		return nil
	}

	if idxT == -5 {
		return &p.TmpM
	}

	if idxT == -6 {
		return nil
	}

	if idxT == -9 {
		return nil
	}

	if idxT == -10 {
		return nil
	}

	// if idxT < -199 {
	// 	if idxT < -249 {
	// 		return &p.CurrentFuncContextM.RegsM.AnysM[(-idxT)-250]
	// 		// } else if idxT < -239 {
	// 		// 	return &((interface{})(p.CurrentFuncContextM.RegsM.StrsM[(-idxT)-240]))
	// 		// } else if idxT < -229 {
	// 		// 	return &p.CurrentFuncContextM.RegsM.FloatsM[(-idxT)-230]
	// 		// } else if idxT < -219 {
	// 		// 	return &p.CurrentFuncContextM.RegsM.IntsM[(-idxT)-220]
	// 		// } else {
	// 		// 	return &p.CurrentFuncContextM.RegsM.CondsM[(-idxT)-210]
	// 	}
	// }

	if idxT < 0 {
		return nil
	}

	// _, ok := p.VarsM[idxT]

	// if !ok {
	// 	return nil
	// }

	contextT := p.CurrentFuncContextM

	nv, ok := contextT.VarsLocalMapM[idxT]

	if !ok {

		for {
			if contextT.Layer < 1 {
				break
			}

			if contextT.Layer < 2 {
				contextT = &p.FuncContextM
			} else {
				contextT = &p.FuncStackM[contextT.Layer-2]
			}

			nv, ok = contextT.VarsLocalMapM[idxT]

			if !ok {
				continue
			}

			return &((*contextT.VarsM)[nv])
		}

		return nil
	}

	return &((*contextT.VarsM)[nv])
	// return &((*contextT.VarsM)[contextT.VarsLocalMapM[idxT]])
}

func (p *XieVM) GetVarRefNative(vA VarRef) interface{} {
	// tk.Pl("def%T %v", vA, vA)

	idxT := vA.Ref

	if idxT == -2 {
		return nil
	}

	if idxT == -3 {
		return nil
	}

	if idxT == -8 {
		return nil
	}

	if idxT == -7 {
		return nil
	}

	if idxT == -5 {
		return tk.GetRef(&p.TmpM)
		// tmpv := p.TmpM

		// switch nnv := tmpv.(type) {
		// default:
		// 	tk.Pl("%T %v", nnv, nnv)
		// 	return &nnv
		// }
		// // tk.Pl("def%T %v", tmpv, tmpv)

		// return &p.TmpM
	}

	if idxT == -6 {
		return nil
	}

	if idxT == -9 {
		return nil
	}

	if idxT == -10 {
		return nil
	}

	if idxT < 0 {
		return nil
	}

	contextT := p.CurrentFuncContextM

	nv, ok := contextT.VarsLocalMapM[idxT]

	if !ok {

		for {
			if contextT.Layer < 1 {
				break
			}

			if contextT.Layer < 2 {
				contextT = &p.FuncContextM
			} else {
				contextT = &p.FuncStackM[contextT.Layer-2]
			}

			nv, ok = contextT.VarsLocalMapM[idxT]

			if !ok {
				continue
			}

			return tk.GetRef(&(*contextT.VarsM)[nv])

			// tmpv := (*contextT.VarsM)[nv]

			// switch nnv := tmpv.(type) {
			// case bool:
			// 	return &nnv
			// case byte:
			// 	return &nnv
			// case rune:
			// 	return &nnv
			// case int:
			// 	return &nnv
			// case int64:
			// 	return &nnv
			// case float32:
			// 	return &nnv
			// case float64:
			// 	return &nnv
			// case string:
			// 	return &nnv
			// case interface{}:
			// 	return &nnv
			// case []byte:
			// 	return &nnv
			// case []rune:
			// 	return &nnv
			// case []int:
			// 	return &nnv
			// case []int64:
			// 	return &nnv
			// case []float32:
			// 	return &nnv
			// case []float64:
			// 	return &nnv
			// case []string:
			// 	return &nnv
			// case map[string]string:
			// 	return &nnv
			// case map[string]interface{}:
			// 	return &nnv
			// case bytes.Buffer:
			// 	return &nnv
			// case strings.Builder:
			// 	return &nnv
			// default:
			// 	return &nnv
			// }
		}

		return nil
	}

	// tk.Pl("def2: %T %v", nnv, nnv)

	return tk.GetRef(&(*contextT.VarsM)[nv])

	// tmpv := (*contextT.VarsM)[nv]
	// switch nnv := tmpv.(type) {
	// case bool:
	// 	return &nnv
	// case byte:
	// 	return &nnv
	// case rune:
	// 	return &nnv
	// case int:
	// 	return &nnv
	// case int64:
	// 	return &nnv
	// case float32:
	// 	return &nnv
	// case float64:
	// 	return &nnv
	// case string:
	// 	return &nnv
	// case interface{}:
	// 	return &nnv
	// case []byte:
	// 	return &nnv
	// case []rune:
	// 	return &nnv
	// case []int:
	// 	return &nnv
	// case []int64:
	// 	return &nnv
	// case []float32:
	// 	return &nnv
	// case []float64:
	// 	return &nnv
	// case []string:
	// 	return &nnv
	// case map[string]string:
	// 	return &nnv
	// case map[string]interface{}:
	// 	return &nnv
	// case strings.Builder:
	// 	return &nnv
	// case bytes.Buffer:
	// 	return &nnv

	// default:
	// 	tk.Pl("def2: %T %v", nnv, nnv)
	// 	return &nnv
	// }

	// return nil
	// return &((*contextT.VarsM)[contextT.VarsLocalMapM[idxT]])
}

func (p *XieVM) GetVarValueGlobal(vA VarRef) interface{} {
	idxT := vA.Ref

	if idxT == -2 {
		return Undefined
	}

	if idxT == -3 {
		return vA.Value
	}

	if idxT == -5 {
		return p.TmpM
	}

	if idxT == -8 {
		return p.Pop()
	}

	if idxT == -7 {
		return p.Peek()
	}

	if idxT == -6 {
		return Undefined
	}

	if idxT == -1 { // $debug
		return tk.ToJSONX(p, "-indent", "-sort")
	}

	if idxT == -9 {
		return p.EvalExpression(vA.Value.(string))
	}

	if idxT == -10 {
		return p.QuickEval(vA.Value.(string))
	}

	ifUnrefT := false
	if idxT == -12 { // unref
		ifUnrefT = true

		idxT = tk.ToInt(vA.Value, -99)
	}

	ifRefT := false
	if idxT == -15 { // ref
		ifRefT = true

		idxT = tk.ToInt(vA.Value, -99)
	}

	if idxT < 0 {
		return Undefined
	}

	contextT := p.FuncContextM

	nv, ok := contextT.VarsLocalMapM[idxT]

	if !ok {
		return Undefined
	}

	if ifUnrefT {
		return tk.GetValue((*contextT.VarsM)[nv])
	}

	if ifRefT {
		return &((*contextT.VarsM)[nv])
	}

	return (*contextT.VarsM)[nv]

	// return p.VarsM[idxT]

	// vT, ok := p.VarsM[idxT]

	// if !ok {
	// 	return Undefined
	// }

	// return vT
}

func (p *XieVM) ParseLine(commandA string) ([]string, string, error) {
	var args []string
	var lineT string

	firstT := true

	// state: 1 - start, quotes - 2, arg - 3
	state := 1
	current := ""
	quote := "`"
	// escapeNext := false

	command := []rune(commandA)

	for i := 0; i < len(command); i++ {
		c := command[i]

		// if escapeNext {
		// 	current += string(c)
		// 	escapeNext = false
		// 	continue
		// }

		// if c == '\\' {
		// 	current += string(c)
		// 	escapeNext = true
		// 	continue
		// }

		if state == 2 {
			if string(c) != quote {
				current += string(c)
			} else {
				current += string(c) // add it

				args = append(args, current)
				if firstT {
					firstT = false
					lineT = string(command[i:])
				}
				current = ""
				state = 1
			}
			continue
		}

		// tk.Pln(string(c), c, c == '`', '`')
		if c == '"' || c == '\'' || c == '`' {
			state = 2
			quote = string(c)

			current += string(c) // add it

			continue
		}

		if state == 3 {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				if firstT {
					firstT = false
					lineT = string(command[i:])
				}
				current = ""
				state = 1
			} else {
				current += string(c)
			}
			// Pl("state: %v, current: %v, args: %v", state, current, args)
			continue
		}

		if c != ' ' && c != '\t' {
			state = 3
			current += string(c)
		}
	}

	if state == 2 {
		return []string{}, lineT, fmt.Errorf("指令中含有未闭合的引号： %v", string(command))
	}

	if current != "" {
		args = append(args, current)
		if firstT {
			firstT = false
			lineT = ""
		}
	}

	return args, lineT, nil
}

func (p *XieVM) NewInstr(codeA string, valuesA *map[string]interface{}) Instr {
	v := strings.TrimSpace(codeA)

	if tk.StartsWith(v, "//") || tk.StartsWith(v, "#") {
		instrT := Instr{Code: 101, ParamLen: 0}
		return instrT
	}

	// var varCountT int

	if tk.StartsWith(v, ":") {
		instrT := Instr{Code: InstrNameSet["pass"], ParamLen: 0}
		return instrT
	}

	listT, lineT, errT := p.ParseLine(v)
	if errT != nil {
		instrT := Instr{Code: InstrNameSet["invalidInstr"], ParamLen: 1, Params: []VarRef{VarRef{Ref: -3, Value: "参数解析失败"}}, Line: lineT}
		return instrT
	}

	lenT := len(listT)

	instrNameT := strings.TrimSpace(listT[0])

	codeT, ok := InstrNameSet[instrNameT]

	if !ok {
		instrT := Instr{Code: InstrNameSet["invalidInstr"], ParamLen: 1, Params: []VarRef{VarRef{Ref: -3, Value: tk.Spr("未知指令：%v", instrNameT)}}}
		return instrT
	}

	instrT := Instr{Code: codeT, Params: make([]VarRef, 0, lenT-1)} //&([]VarRef{})}

	list3T := []VarRef{}

	for j, jv := range listT {
		if j == 0 {
			continue
		}

		if strings.HasPrefix(jv, "~") {
			list3T = append(list3T, VarRef{-3, (*valuesA)[jv]})
		} else {
			list3T = append(list3T, p.ParseVar(jv))
		}

	}

	instrT.Params = append(instrT.Params, list3T...)
	instrT.ParamLen = lenT - 1

	return instrT
}

func (p *XieVM) Load(codeA string) string {

	// originSourceLenT := len(p.SourceM)
	originCodeLenT := len(p.CodeListM)

	sourceT := tk.SplitLines(codeA)

	p.SourceM = append(p.SourceM, sourceT...)

	// p.CodeListM = make([]string, 0, len(p.SourceM))
	// p.InstrListM = make([]Instr, 0, len(p.SourceM))

	// p.LabelsM = make(map[int]int, len(p.SourceM))

	// p.CodeSourceMapM = make(map[int]int, len(p.SourceM))

	pointerT := originCodeLenT

	var varCountT int

	for i := 0; i < len(sourceT); i++ {
		v := strings.TrimSpace(sourceT[i])

		if tk.StartsWith(v, "//") || tk.StartsWith(v, "#") {
			continue
		}

		if tk.StartsWith(v, ":") {
			labelT := strings.TrimSpace(v[1:])

			_, ok := p.VarIndexMapM[labelT]

			if !ok {
				varCountT = len(p.VarIndexMapM)

				p.VarIndexMapM[labelT] = varCountT
				p.VarNameMapM[varCountT] = labelT
			} else {
				return tk.ErrStrf("编译错误(行 %v %v): 重复的标号", i+1, tk.LimitString(p.SourceM[i], 50))
			}

			p.LabelsM[varCountT] = pointerT

			continue
		}

		iFirstT := i
		if tk.Contains(v, "`") {
			if strings.Count(v, "`")%2 != 0 {
				foundT := false
				var j int
				for j = i + 1; j < len(sourceT); j++ {
					if tk.Contains(sourceT[j], "`") {
						v = tk.JoinLines(sourceT[i : j+1])
						foundT = true
						break
					}
				}

				if !foundT {
					return tk.ErrStrf("代码解析错误: ` 未成对(%v)", i)
				}

				i = j
			}
		}

		v = strings.TrimSpace(v)

		if v == "" {
			continue
		}

		p.CodeListM = append(p.CodeListM, v)
		p.CodeSourceMapM[pointerT] = originCodeLenT + iFirstT
		pointerT++
	}

	for i := originCodeLenT; i < len(p.CodeListM); i++ {
		// listT := strings.SplitN(v, " ", 3)
		v := p.CodeListM[i]
		listT, lineT, errT := p.ParseLine(v)
		if errT != nil {
			return p.ErrStrf("参数解析失败：%v", errT)
		}

		lenT := len(listT)

		instrNameT := strings.TrimSpace(listT[0])

		codeT, ok := InstrNameSet[instrNameT]

		if !ok {
			instrT := Instr{Code: codeT, ParamLen: 1, Params: []VarRef{VarRef{Ref: -3, Value: v}}, Line: lineT} //&([]VarRef{})}
			p.InstrListM = append(p.InstrListM, instrT)

			return tk.ErrStrf("编译错误(行 %v/%v %v): 未知指令", i, p.CodeSourceMapM[i]+1, tk.LimitString(p.SourceM[p.CodeSourceMapM[i]], 50))
		}

		instrT := Instr{Code: codeT, Params: make([]VarRef, 0, lenT-1), Line: lineT} //&([]VarRef{})}

		list3T := []VarRef{}

		for j, jv := range listT {
			if j == 0 {
				continue
			}

			list3T = append(list3T, p.ParseVar(jv))
		}

		instrT.Params = append(instrT.Params, list3T...)
		instrT.ParamLen = lenT - 1

		p.InstrListM = append(p.InstrListM, instrT)
	}

	// tk.Plv(p.SourceM)
	// tk.Plv(p.CodeListM)
	// tk.Plv(p.CodeSourceMapM)

	return tk.ToStr(originCodeLenT)
}

func (p *XieVM) PushFunc() {
	// funcContextT := FuncContext{VarsM: make(map[int]interface{}, 10), ReturnPointerM: p.CodePointerM + 1}
	// funcContextT := FuncContext{VarsM: make([]interface{}, 0, 10), VarsLocalMapM: make(map[int]int, 10), ReturnPointerM: p.CodePointerM + 1}
	funcContextT := FuncContext{VarsM: &([]interface{}{}), VarsLocalMapM: make(map[int]int, 10), ReturnPointerM: p.CodePointerM + 1, Layer: p.FuncStackPointerM + 1, DeferStackM: tk.NewSimpleStack()}

	lenT := len(p.FuncStackM)

	if p.FuncStackPointerM >= lenT {
		p.FuncStackM = append(p.FuncStackM, funcContextT)
	} else {
		p.FuncStackM[p.FuncStackPointerM] = funcContextT
	}

	p.CurrentFuncContextM = &(p.FuncStackM[p.FuncStackPointerM])

	p.FuncStackPointerM++

	// p.CurrentFuncContextM.RegsM = &(p.FuncStackM[len(p.FuncStackM)-1].RegsM)
	// p.CurrentVarsM = &(p.FuncStackM[len(p.FuncStackM)-1].VarsM)

}

func (p *XieVM) PopFunc() int {
	if p.FuncStackPointerM < 1 {
		return 0
	}

	p.FuncStackPointerM--
	funcContextT := p.FuncStackM[p.FuncStackPointerM]

	if p.FuncStackPointerM < 1 {
		// p.CurrentFuncContextM.RegsM = &(p.RegsM)
		// p.CurrentVarsM = &(p.VarsM)
		p.CurrentFuncContextM = &(p.FuncContextM)
	} else {
		p.CurrentFuncContextM = &(p.FuncStackM[p.FuncStackPointerM-1])
		// p.CurrentFuncContextM.RegsM = &(p.FuncStackM[len(p.FuncStackM)-1].RegsM)
		// p.CurrentVarsM = &(p.FuncStackM[len(p.FuncStackM)-1].VarsM)

	}

	return funcContextT.ReturnPointerM
}

func (p *XieVM) SetVarInt(keyA int, vA interface{}) error {
	// if p.FuncContextM.VarsM == nil {
	// 	p.InitVM()
	// }

	if keyA == -2 { // 丢弃
		return nil
	}

	// if keyA < -199 {
	// 	if keyA < -249 {
	// 		p.CurrentFuncContextM.RegsM.AnysM[(-keyA)-250] = vA
	// 	} else if keyA < -239 {
	// 		p.CurrentFuncContextM.RegsM.StrsM[(-keyA)-240] = vA.(string)
	// 	} else if keyA < -229 {
	// 		p.CurrentFuncContextM.RegsM.FloatsM[(-keyA)-230] = vA.(float64)
	// 	} else if keyA < -219 {
	// 		p.CurrentFuncContextM.RegsM.IntsM[(-keyA)-220] = vA.(int)
	// 	} else {
	// 		p.CurrentFuncContextM.RegsM.CondsM[(-keyA)-210] = vA.(bool)
	// 	}

	// 	return nil
	// }

	if keyA == -6 {
		p.Push(vA)
		return nil
	}

	if keyA == -5 {
		p.TmpM = vA
		return nil
	}

	if keyA == -4 {
		fmt.Println(vA)
		return nil
	}

	if keyA < 0 {
		return fmt.Errorf("无效的变量索引")
	}

	contextT := p.CurrentFuncContextM

	localIdxT, ok := contextT.VarsLocalMapM[keyA]

	if !ok {
		contextTmpT := contextT
		for {
			if contextTmpT.Layer < 1 {
				break
			}

			if contextTmpT.Layer < 2 {
				contextTmpT = &p.FuncContextM
			} else {
				contextTmpT = &p.FuncStackM[contextTmpT.Layer-2]
			}

			localIdxT, ok = contextTmpT.VarsLocalMapM[keyA]

			if !ok {
				continue
			}

			(*contextTmpT.VarsM)[localIdxT] = vA

			return nil
		}

		// return Undefined

		// return (*contextT.VarsM)[nv]

		localIdxT = len((*contextT.VarsM))

		contextT.VarsLocalMapM[keyA] = localIdxT

		(*contextT.VarsM) = append((*contextT.VarsM), vA)

		// tk.Pln(contextT.VarsM, "***")

		return nil
	}

	// tk.Pln(contextT.VarsLocalMapM, contextT.VarsM, keyA, localIdxT, ok)
	(*contextT.VarsM)[localIdxT] = vA
	// varsT := *(p.CurrentVarsM)

	// idxT, ok := varsT[keyA]

	// varsT[keyA] = vA

	return nil
}

func (p *XieVM) SetVarIntLocal(keyA int, vA interface{}) error {
	if keyA == -6 {
		p.Push(vA)
		return nil
	}

	if keyA == -5 {
		p.TmpM = vA
		return nil
	}

	if keyA == -4 {
		fmt.Println(vA)
		return nil
	}

	if keyA == -2 { // 丢弃
		return nil
	}

	// if keyA < -199 {
	// 	if keyA < -249 {
	// 		p.CurrentFuncContextM.RegsM.AnysM[(-keyA)-250] = vA
	// 	} else if keyA < -239 {
	// 		p.CurrentFuncContextM.RegsM.StrsM[(-keyA)-240] = vA.(string)
	// 	} else if keyA < -229 {
	// 		p.CurrentFuncContextM.RegsM.FloatsM[(-keyA)-230] = vA.(float64)
	// 	} else if keyA < -219 {
	// 		p.CurrentFuncContextM.RegsM.IntsM[(-keyA)-220] = vA.(int)
	// 	} else {
	// 		p.CurrentFuncContextM.RegsM.CondsM[(-keyA)-210] = vA.(bool)
	// 	}

	// 	return nil
	// }

	if keyA < 0 {
		return fmt.Errorf("无效的变量索引")
	}

	contextT := p.CurrentFuncContextM

	localIdxT, ok := contextT.VarsLocalMapM[keyA]

	if !ok {
		localIdxT = len((*contextT.VarsM))

		contextT.VarsLocalMapM[keyA] = localIdxT

		(*contextT.VarsM) = append((*contextT.VarsM), vA)

		// tk.Pln(contextT.VarsM, "***")

		return nil
	}

	// tk.Pln(contextT.VarsLocalMapM, contextT.VarsM, keyA, localIdxT, ok)
	(*contextT.VarsM)[localIdxT] = vA
	// varsT := *(p.CurrentVarsM)

	// idxT, ok := varsT[keyA]

	// varsT[keyA] = vA

	return nil
}

func (p *XieVM) SetVarIntGlobal(keyA int, vA interface{}) error {
	// if p.FuncContextM.VarsM == nil {
	// 	p.InitVM()
	// }

	if keyA == -2 { // 丢弃
		return nil
	}

	if keyA == -6 {
		p.Push(vA)
		return nil
	}

	if keyA == -5 {
		p.TmpM = vA
		return nil
	}

	if keyA == -4 {
		fmt.Println(vA)
		return nil
	}

	// if keyA < -199 {
	// 	if keyA < -249 {
	// 		p.CurrentFuncContextM.RegsM.AnysM[(-keyA)-250] = vA
	// 	} else if keyA < -239 {
	// 		p.CurrentFuncContextM.RegsM.StrsM[(-keyA)-240] = vA.(string)
	// 	} else if keyA < -229 {
	// 		p.CurrentFuncContextM.RegsM.FloatsM[(-keyA)-230] = vA.(float64)
	// 	} else if keyA < -219 {
	// 		p.CurrentFuncContextM.RegsM.IntsM[(-keyA)-220] = vA.(int)
	// 	} else {
	// 		p.CurrentFuncContextM.RegsM.CondsM[(-keyA)-210] = vA.(bool)
	// 	}

	// 	return nil
	// }

	if keyA < 0 {
		return fmt.Errorf("无效的变量编号")
	}

	// contextT := p.FuncContextM

	localIdxT, ok := p.FuncContextM.VarsLocalMapM[keyA]

	if !ok {
		localIdxT = len(*p.FuncContextM.VarsM)

		p.FuncContextM.VarsLocalMapM[keyA] = localIdxT

		*p.FuncContextM.VarsM = append(*p.FuncContextM.VarsM, vA)

		return nil
	}

	(*p.FuncContextM.VarsM)[localIdxT] = vA
	// p.VarsM[keyA] = vA

	return nil
}

func (p *XieVM) SetVar(keyA string, vA interface{}) error {
	// if p.FuncContextM.VarsM == nil {
	// 	p.InitVM()
	// }

	idxT, ok := p.VarIndexMapM[keyA]
	// tk.Pln(keyA, idxT, ok, p.VarIndexMapM)

	if !ok {
		idxT = len(p.VarIndexMapM)

		p.VarIndexMapM[keyA] = idxT
		p.VarNameMapM[idxT] = keyA

		// tk.Pln(idxT, p.VarIndexMapM, p.VarNameMapM)
	}

	contextT := p.CurrentFuncContextM

	localIdxT, ok := contextT.VarsLocalMapM[idxT]

	// tk.Pln(idxT, localIdxT, ok, contextT.VarsLocalMapM)

	if !ok {
		contextTmpT := contextT
		for {
			if contextTmpT.Layer < 1 {
				break
			}

			if contextTmpT.Layer < 2 {
				contextTmpT = &p.FuncContextM
			} else {
				contextTmpT = &p.FuncStackM[contextTmpT.Layer-2]
			}

			localIdxT, ok = contextTmpT.VarsLocalMapM[idxT]

			if !ok {
				continue
			}

			(*contextTmpT.VarsM)[localIdxT] = vA

			return nil
		}

		localIdxT = len((*contextT.VarsM))

		contextT.VarsLocalMapM[idxT] = localIdxT

		(*contextT.VarsM) = append((*contextT.VarsM), vA)

		// tk.Pln(idxT, localIdxT, contextT.VarsLocalMapM, contextT.VarsM, "---")

		return nil
	}

	(*contextT.VarsM)[localIdxT] = vA

	// lenT := len(contextT.VarsM)

	// if lenT < idxT {

	// }

	// varsT := *(p.CurrentVarsM)

	// if len(varsT) < lenT {
	// 	varsT = append(varsT, vA)
	// }

	// ()[lenT] = vA
	return nil
}

func (p *XieVM) SetVarGlobal(keyA string, vA interface{}) {
	// if p.FuncContextM.VarsM == nil {
	// 	p.InitVM()
	// }

	idxT, ok := p.VarIndexMapM[keyA]

	if !ok {
		lenT := len(p.VarIndexMapM) + 1

		p.VarIndexMapM[keyA] = lenT
		p.VarNameMapM[lenT] = keyA

	}

	// contextT := *(p.CurrentFuncContextM)

	localIdxT, ok := p.FuncContextM.VarsLocalMapM[idxT]

	if !ok {
		localIdxT = len(*p.FuncContextM.VarsM)

		p.FuncContextM.VarsLocalMapM[idxT] = localIdxT

		*p.FuncContextM.VarsM = append(*p.FuncContextM.VarsM, vA)

		return
	}

	(*p.FuncContextM.VarsM)[localIdxT] = vA
	// lenT := len(p.VarIndexMapM) + 100

	// p.VarIndexMapM[keyA] = lenT + 1
	// p.VarNameMapM[lenT+1] = keyA

	// p.VarsM[lenT+1] = vA
}

func (p *XieVM) PushVar(vA interface{}) {
	// if p.FuncContextM.VarsM == nil {
	// 	p.InitVM()
	// }

	p.Push(vA)
}

func (p *XieVM) GetVarInt(keyA int) interface{} {
	// if p.FuncContextM.VarsM == nil {
	// 	p.InitVM()
	// }

	contextT := *(p.CurrentFuncContextM)

	localIdxT, ok := contextT.VarsLocalMapM[keyA]

	if !ok {
		return Undefined
	}

	return (*contextT.VarsM)[localIdxT]
}

func (p *XieVM) GetVar(keyA string) interface{} {
	// if p.FuncContextM.VarsM == nil {
	// 	p.InitVM()
	// }

	idxT, ok := p.VarIndexMapM[keyA]

	if !ok {
		return Undefined

	}

	contextT := *(p.CurrentFuncContextM)

	localIdxT, ok := contextT.VarsLocalMapM[idxT]

	if !ok {
		return Undefined
	}

	return (*contextT.VarsM)[localIdxT]

	// lenT := len(p.FuncStackM)

	// if lenT > 0 {
	// 	for i := lenT - 1; i >= 0; i-- {
	// 		varsT := p.FuncStackM[i].VarsM

	// 		vT, ok := varsT[keyA]

	// 		if ok {
	// 			return vT
	// 		}
	// 	}
	// }

	// return p.VarsM[keyA]
}

// get current vars in context
func (p *XieVM) GetVars() []interface{} {
	// if p.FuncContextM.VarsM == nil {
	// 	p.InitVM()
	// }

	// lenT := len(p.FuncStackM)

	// if lenT > 0 {
	// 	return *p.FuncStackM[lenT-1].VarsM
	// }

	if p.FuncStackPointerM < 1 {
		return *p.FuncContextM.VarsM
	}

	return *p.FuncStackM[p.FuncStackPointerM-1].VarsM
	// return *p.FuncContextM.VarsM
}

// func (p *XieVM) GetRegs() *Regs {
// 	// lenT := len(p.FuncStackM)

// 	if p.FuncStackPointerM > 0 {
// 		return (p.FuncStackM[p.FuncStackPointerM-1].RegsM)
// 	}

// 	// if lenT > 0 {
// 	// 	return (p.FuncStackM[lenT-1].RegsM)
// 	// }

// 	return (p.FuncContextM.RegsM)
// }

func (p *XieVM) Push(vA interface{}) {
	lenT := len(p.StackM)

	if p.StackPointerM >= lenT {
		p.StackM = append(p.StackM, vA)
	} else {
		p.StackM[p.StackPointerM] = vA
	}

	p.StackPointerM++
}

// func (p *XieVM) PushLocal(vA interface{}) {
// 	if p.StackM == nil {
// 		p.InitVM()
// 	}

// 	lenT := len(p.FuncStackM)

// 	if lenT > 0 {
// 		p.FuncStackM[lenT-1].StackM = append(p.FuncStackM[lenT-1].StackM, vA)
// 		return
// 	}

// 	p.StackM = append(p.StackM, vA)
// }

func (p *XieVM) Pop() interface{} {
	if p.StackPointerM < 1 {
		return Undefined
	}

	p.StackPointerM--
	rs := p.StackM[p.StackPointerM]

	return rs
}

// func (p *XieVM) PopLocal() interface{} {
// 	if p.StackM == nil {
// 		p.InitVM()

// 		return Undefined
// 	}

// 	len1T := len(p.FuncStackM)

// 	if len1T > 0 {
// 		lenT := len(p.FuncStackM[len1T-1].StackM)

// 		if lenT < 1 {
// 			return Undefined
// 		}

// 		rs := p.FuncStackM[len1T-1].StackM[lenT-1]

// 		p.FuncStackM[len1T-1].StackM = p.FuncStackM[len1T-1].StackM[0 : lenT-1]

// 		return rs
// 	}

// 	lenT := len(p.StackM)

// 	if lenT < 1 {
// 		return Undefined
// 	}

// 	rs := p.StackM[lenT-1]

// 	p.StackM = p.StackM[0 : lenT-1]

// 	return rs
// }

// func (p *XieVM) PeekLocal() interface{} {
// 	if p.StackM == nil {
// 		p.InitVM()

// 		return Undefined
// 	}

// 	len1T := len(p.FuncStackM)

// 	if len1T > 0 {
// 		lenT := len(p.FuncStackM[len1T-1].StackM)

// 		if lenT < 1 {
// 			return Undefined
// 		}

// 		rs := p.FuncStackM[len1T-1].StackM[lenT-1]

// 		return rs
// 	}

// 	lenT := len(p.StackM)

// 	if lenT < 1 {
// 		return Undefined
// 	}

// 	rs := p.StackM[lenT-1]

// 	return rs
// }

func (p *XieVM) Pops() string {
	if p.StackPointerM < 1 {
		return tk.ErrStrf("no value")
	}

	p.StackPointerM--
	rs := p.StackM[p.StackPointerM]

	return tk.ToStr(rs)
}

func (p *XieVM) Peek() interface{} {
	if p.StackPointerM < 1 {
		return Undefined
	}

	// p.StackPointerM--
	rs := p.StackM[p.StackPointerM-1]

	return rs
}

// func (p *XieVM) GetName(nameA string) string {
// 	if tk.StartsWith(nameA, "$") {
// 		return nameA[1:]
// 	} else {
// 		return nameA
// 	}
// }

func (p *XieVM) ParamsToStrs(v Instr, fromA int) []string {

	lenT := len(v.Params)

	sl := make([]string, 0, lenT)

	for i := fromA; i < lenT; i++ {
		sl = append(sl, tk.ToStr(p.GetVarValue(v.Params[i])))
	}

	return sl
}

func (p *XieVM) ParamsToList(v Instr, fromA int) []interface{} {

	lenT := len(v.Params)

	sl := make([]interface{}, 0, lenT)

	for i := fromA; i < lenT; i++ {
		sl = append(sl, p.GetVarValue(v.Params[i]))
	}

	return sl
}

func (p *XieVM) ParamsToReflectValueList(v Instr, fromA int) []reflect.Value {

	lenT := len(v.Params)

	sl := make([]reflect.Value, 0, lenT)

	for i := fromA; i < lenT; i++ {
		sl = append(sl, reflect.ValueOf(p.GetVarValue(v.Params[i])))
	}

	return sl
}

func (p *XieVM) ParamsToRefList(v *Instr, fromA int) []interface{} {

	lenT := len(v.Params)

	sl := make([]interface{}, 0, lenT)

	for i := fromA; i < lenT; i++ {
		// switch nv := v.Params[i].Value.(type) {
		// case byte:
		// 	sl = append(sl, &(v.Params[i].Value.(byte)))
		// }
		sl = append(sl, &(v.Params[i].Value))
	}

	return sl
}

func (p *XieVM) ErrStrf(formatA string, argsA ...interface{}) string {
	// tk.Pl("dbg: %v", tk.ToJSONX(p, "-sort"))
	if p.VerboseM {
		tk.Pl(fmt.Sprintf("TXERROR:(Line %v: %v) ", p.CodeSourceMapM[p.CodePointerM]+1, tk.LimitString(p.SourceM[p.CodeSourceMapM[p.CodePointerM]], 50))+formatA, argsA...)
	}

	return fmt.Sprintf(fmt.Sprintf("TXERROR:(Line %v: %v) ", p.CodeSourceMapM[p.CodePointerM]+1, tk.LimitString(p.SourceM[p.CodeSourceMapM[p.CodePointerM]], 50))+formatA, argsA...)
}

func (p *XieVM) Debug() {
	tk.Pln(tk.ToJSONX(p, "-indent", "-sort"))
}

func help(wordA string) {
	switch wordA {
	case "":
		tk.Pln(`谢语言版本（Xielang ver.）` + VersionG + `

用法：

	xie -- 直接执行谢语言交互式命令行环境

	xie c:\test.xie -- 执行指定路径的谢语言代码文件

	xie http://example.com/test.xie -- 执行指定网络地址的谢语言脚本文件

在交互式命令行环境中：

	-- 输入help cmds查看指令列表，help objects查看内置对象列表

	-- 输入quit或按Ctrl-C退出
`)
	default:
		tk.Pln("未知的帮助条目，输入help cmds查看指令列表，help objects查看内置对象列表")
	}
}

var memberMapG = map[string]map[string]interface{}{
	"": map[string]interface{}{
		"toStr": tk.ToStr,
	},
	"string": map[string]interface{}{
		"trimSpace": strings.TrimSpace,
	},
}

func callGoFunc(funcA interface{}, thisA interface{}, argsA ...interface{}) interface{} {

	switch nv := funcA.(type) {
	case func(string) string:
		rsT := nv(thisA.(string))
		return rsT
	case func(interface{}) string:
		rsT := nv(thisA)
		return rsT
	default:
		tk.Pl("nv: %v, type: %T, this: %v", nv, funcA, thisA)
	}

	return nil
}

var leBufG []string
var leLineEndG string = "\n"
var leSilentG bool = true
var leSSHInfoG map[string]string = map[string]string{}

func SetLeVSilent(vA bool) {
	leSilentG = vA
}

func leClear() {
	leBufG = make([]string, 0, 100)
}

func leLoadString(strA string) {
	if leBufG == nil {
		leClear()
	}

	leBufG = tk.SplitLines(strA)
}

func leSaveString() string {
	if leBufG == nil {
		leClear()
	}

	return tk.JoinLines(leBufG, leLineEndG)
}

func leLoadFile(fileNameA string) error {
	if leBufG == nil {
		leClear()
	}

	strT, errT := tk.LoadStringFromFileE(fileNameA)

	if errT != nil {
		return errT
	}

	if strings.Contains(strT, "\r") {
		leLineEndG = "\r\n"
	} else {
		leLineEndG = "\n"
	}

	leBufG = tk.SplitLines(strT)
	// leBufG, errT = tk.LoadStringListBuffered(fileNameA, false, false)

	return nil
}

func leLoadUrl(urlA string) error {
	if leBufG == nil {
		leClear()
	}

	strT := tk.DownloadWebPageX(urlA)

	if tk.IsErrStr(strT) {
		return tk.ErrStrToErr(strT)
	}

	leBufG = tk.SplitLines(strT)
	// leBufG, errT = tk.LoadStringListBuffered(fileNameA, false, false)

	return nil
}

func leSaveFile(fileNameA string) error {
	if leBufG == nil {
		leClear()
	}

	var errT error

	textT := tk.JoinLines(leBufG, leLineEndG)

	if tk.IsErrStr(textT) {
		return tk.Errf(tk.GetErrStr(textT))
	}

	errT = tk.SaveStringToFileE(textT, fileNameA)

	return errT
}

func leLoadClip() error {
	if leBufG == nil {
		leClear()
	}

	textT := tk.GetClipText()

	if tk.IsErrStr(textT) {
		return tk.Errf(tk.GetErrStr(textT))
	}

	leBufG = tk.SplitLines(textT)

	return nil
}

func leSaveClip() error {
	if leBufG == nil {
		leClear()
	}

	textT := tk.JoinLines(leBufG, leLineEndG)

	if tk.IsErrStr(textT) {
		return tk.Errf(tk.GetErrStr(textT))
	}

	return tk.SetClipText(textT)
}

func leViewAll(argsA ...string) error {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return tk.Errf("buffer not initalized")
	}

	if tk.IfSwitchExistsWhole(argsA, "-nl") {
		textT := tk.JoinLines(leBufG, leLineEndG)

		tk.Pln(textT)

	} else {
		for i, v := range leBufG {
			tk.Pl("%v: %v", i, v)
		}
	}

	return nil
}

func leViewLine(idxA int) error {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return tk.Errf("buffer not initalized")
	}

	if idxA < 0 || idxA >= len(leBufG) {
		return tk.Errf("line index out of range")
	}

	tk.Pln(leBufG[idxA])

	return nil
}

func leSort(descentA bool) error {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return tk.Errf("buffer not initalized")
	}

	if descentA {
		sort.Sort(sort.Reverse(sort.StringSlice(leBufG)))
	} else {
		sort.Sort(sort.StringSlice(leBufG))
	}

	return nil
}

func leConvertToUTF8(srcEncA ...string) error {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return tk.Errf("buffer not initalized")
	}

	encT := ""

	if len(srcEncA) > 0 {
		encT = srcEncA[0]
	}

	leBufG = tk.SplitLines(tk.ConvertStringToUTF8(tk.JoinLines(leBufG, leLineEndG), encT))

	return nil
}

func leLineEnd(lineEndA ...string) string {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return tk.ErrStrf("buffer not initalized")
	}

	if len(lineEndA) > 0 {
		leLineEndG = lineEndA[0]
	} else {
		return leLineEndG
	}

	return ""
}

func leSilent(silentA ...bool) bool {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return false
	}

	if len(silentA) > 0 {
		leSilentG = silentA[0]
		return leSilentG
	}

	return leSilentG
}

func leFind(regA string) []string {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return nil
	}

	aryT := []string{}

	for i, v := range leBufG {
		if tk.RegContains(v, regA) {
			aryT = append(aryT, fmt.Sprintf("%v: %v", i, v))
		}
	}

	return aryT
}

func leReplace(regA string, replA string) []string {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return nil
	}

	aryT := []string{}

	for i, v := range leBufG {
		if tk.RegContains(v, regA) {
			leBufG[i] = tk.RegReplace(v, regA, replA)
			aryT = append(aryT, fmt.Sprintf("%v: %v -> %v", i, v, leBufG[i]))
		}
	}

	return aryT
}

func leGetLine(idxA int) string {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return tk.ErrStrf("buffer not initalized")
	}

	if idxA < 0 || idxA >= len(leBufG) {
		return tk.ErrStrf("line index out of range")
	}

	return leBufG[idxA]
}

func leSetLine(idxA int, strA string) error {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return tk.Errf("buffer not initalized")
	}

	if idxA < 0 || idxA >= len(leBufG) {
		return tk.Errf("line index out of range")
	}

	leBufG[idxA] = strA

	return nil
}

func leSetLines(startA int, endA int, strA string) error {
	if leBufG == nil {
		leClear()
	}

	if startA > endA {
		return tk.Errf("start index greater than end index")
	}

	listT := tk.SplitLines(strA)

	if endA < 0 {
		rs := make([]string, 0, len(leBufG)+len(listT))

		rs = append(rs, listT...)
		rs = append(rs, leBufG...)

		leBufG = rs

		return nil
	}

	if startA >= len(leBufG) {
		leBufG = append(leBufG, listT...)

		return nil
	}

	if startA < 0 {
		startA = 0
	}

	if endA >= len(leBufG) {
		endA = len(leBufG) - 1
	}

	rs := make([]string, 0, len(leBufG)+len(listT)-1)

	rs = append(rs, leBufG[:startA]...)
	rs = append(rs, listT...)
	rs = append(rs, leBufG[endA+1:]...)

	leBufG = rs

	return nil
}

func leInsertLine(idxA int, strA string) error {
	if leBufG == nil {
		leClear()
	}

	// if leBufG == nil {
	// 	return tk.Errf("buffer not initalized")
	// }

	// if idxA < 0 || idxA >= len(leBufG) {
	// 	return tk.Errf("line index out of range")
	// }

	if idxA < 0 {
		idxA = 0
	}

	listT := tk.SplitLines(strA)

	if idxA >= len(leBufG) {
		leBufG = append(leBufG, listT...)
	} else {
		rs := make([]string, 0, len(leBufG)+1)

		rs = append(rs, leBufG[:idxA]...)
		rs = append(rs, listT...)
		rs = append(rs, leBufG[idxA:]...)

		leBufG = rs

	}

	return nil
}

func leAppendLine(strA string) error {
	if leBufG == nil {
		leClear()
	}

	// if leBufG == nil {
	// 	return tk.Errf("buffer not initalized")
	// }

	// if idxA < 0 || idxA >= len(leBufG) {
	// 	return tk.Errf("line index out of range")
	// }

	listT := tk.SplitLines(strA)

	leBufG = append(leBufG, listT...)

	return nil
}

func leRemoveLine(idxA int) error {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return tk.Errf("buffer not initalized")
	}

	if idxA < 0 || idxA >= len(leBufG) {
		return tk.Errf("line index out of range")
	}

	rs := make([]string, 0, len(leBufG)+1)

	rs = append(rs, leBufG[:idxA]...)
	rs = append(rs, leBufG[idxA+1:]...)

	leBufG = rs

	return nil
}

func leRemoveLines(startA int, endA int) error {
	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return tk.Errf("buffer not initalized")
	}

	if startA < 0 || startA >= len(leBufG) {
		return tk.Errf("start line index out of range")
	}

	if endA < 0 || endA >= len(leBufG) {
		return tk.Errf("end line index out of range")
	}

	if startA > endA {
		return tk.Errf("start line index greater than end line index")
	}

	rs := make([]string, 0, len(leBufG)+1)

	rs = append(rs, leBufG[:startA]...)
	rs = append(rs, leBufG[endA+1:]...)

	leBufG = rs

	return nil
}

func (p *XieVM) RunLine(lineA int, codeA ...Instr) (resultR interface{}) {
	defer func() {
		if r := recover(); r != nil {
			// tk.Printfln("exception: %v", r)

			if p.ErrorHandlerM > -1 {
				// tk.Pln()

				p.Push(p.ErrStrf("运行时异常: %v\n%v", r, string(debug.Stack())))
				p.Push(tk.ToStr(r))
				p.Push(p.CodeSourceMapM[p.CodePointerM] + 1)
				resultR = p.ErrorHandlerM
				return
			}

			// tk.Pln(p.ErrStrf("运行时异常: %v\n%v", r, string(debug.Stack())))
			// tk.Exit()
			resultR = p.ErrStrf("运行时异常: %v\n%v", r, string(debug.Stack()))

			return
		}
	}()

	if lineA >= len(p.InstrListM) {
		return p.ErrStrf("无效的代码序号: %v/%v", lineA, len(p.InstrListM))
	}

	var instrT Instr

	if len(codeA) > 0 {
		instrT = codeA[0]
	} else {
		instrT = p.InstrListM[lineA]
	}

	if p.VerbosePlusM {
		tk.Plv(instrT)
	}

	cmdT := instrT.Code

	switch cmdT {
	case 12: // invalidInstr
		return p.ErrStrf("无效的指令：%v", instrT.Params[0].Value)
	case 100: // version
		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		p.SetVarInt(pr, VersionG)

		return ""
	case 101: // pass
		return ""
	case 102: // debug
		if instrT.ParamLen > 0 {
			returnStrT := ""

			if instrT.ParamLen > 1 {
				lastParamT := instrT.Params[instrT.ParamLen-1]
				if lastParamT.Ref == -3 {
					tmps := lastParamT.Value.(string)
					if tmps == "exit" || tmps == "终止" {
						returnStrT = "exit"
					}
				}
			}

			if instrT.Params[0].Ref == -3 {
				vTypeT := instrT.Params[0].Value.(string)
				if vTypeT == "exit" || vTypeT == "终止" {
					tk.Pln(tk.ToJSONX(p, "-sort"))
					return returnStrT
				} else if vTypeT == "context" || vTypeT == "上下文" {
					tk.Pl("[行：%v] 堆栈: %v, 上下文: %v", p.CodeSourceMapM[p.CodePointerM]+1, p.StackM[:p.StackPointerM], tk.ToJSONX(p.CurrentFuncContextM, "-sort"))
					return returnStrT
				} else if vTypeT == "contexts" || vTypeT == "全上下文" {
					tk.Pl("[行：%v] 堆栈: %v, 主上下文: %v, 函数栈: %v, 当前上下文: %v", p.CodeSourceMapM[p.CodePointerM]+1, p.StackM[:p.StackPointerM], tk.ToJSONX(p.FuncContextM, "-sort"), tk.ToJSONX(p.FuncStackM[:p.FuncStackPointerM], "-sort"), tk.ToJSONX(p.CurrentFuncContextM, "-sort"))
					return returnStrT
				}
			} else if instrT.Params[0].Ref >= 0 {
				v1, layerT := p.GetVarValueWithLayer(instrT.Params[0])

				tk.Pl("[变量信息]: %v(L: %v) -> (%T) %v", instrT.Params[0].Ref, layerT, v1, v1)
			}
		} else {
			tmps := tk.ToJSONX(p, "-sort")

			if tk.IsErrStr(tmps) {
				tmps = tk.Sdump(p)
			}

			tk.Pln(tmps)
		}

		return ""
	case 103: // debugInfo
		pr := -5
		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		p.SetVarInt(pr, tk.ToJSONX(p, "-indent", "-sort"))

		return ""
	case 104: // varInfo
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0
		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1, layerT := p.GetVarValueWithLayer(instrT.Params[v1p])

		p.SetVarInt(pr, fmt.Sprintf("[变量信息]: %v(行：%v) -> (%T) %v", instrT.Params[v1p].Ref, layerT, v1, v1))

		return ""
	case 105: // help
		wordT := ""
		if instrT.ParamLen > 0 {
			wordT = tk.ToStr(p.GetVarValue(instrT.Params[0]))
		}

		help(wordT)

		return ""
	case 106: // onError
		if instrT.ParamLen < 1 {
			p.ErrorHandlerM = -1
			return ""
		}

		p.ErrorHandlerM = tk.ToInt(p.GetVarValue(instrT.Params[0]))

		return ""
	case 107: // dumpf
		if instrT.ParamLen < 1 {
			tk.Dump(p)
			return ""
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[0]))

		if v1 == "all" {
			tk.Dump(p)
		} else if v1 == "vars" {
			for k, v := range p.FuncContextM.VarsLocalMapM {
				tk.Dumpf("%v -> %v", p.VarNameMapM[k], (*(p.FuncContextM.VarsM))[v])
			}

		} else if v1 == "labels" {
			for k, v := range p.LabelsM {
				tk.Dumpf("%v -> %v/%v (%v)", p.VarNameMapM[k], v, p.CodeSourceMapM[v], tk.LimitString(p.SourceM[p.CodeSourceMapM[v]], 50))
			}

		} else {
			tk.Dumpf(v1, p.ParamsToList(instrT, 1)...)
		}

		return ""
	case 109: // defer
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[0]))

		codeT, ok := InstrNameSet[v1]

		if !ok {
			return p.ErrStrf("未知指令：%v", v1)
		}

		instrT := Instr{Code: codeT, Params: instrT.Params[1:], ParamLen: instrT.ParamLen - 1, Line: tk.RemoveFirstSubString(strings.TrimSpace(instrT.Line), v1)} //&([]VarRef{})}

		p.CurrentFuncContextM.DeferStackM.Push(instrT)

		return ""
	case 111: // isUndef
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		p.SetVarInt(pr, v1 == Undefined)

		return ""

	case 112: // isDef
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		p.SetVarInt(pr, v1 != Undefined)

		return ""

	case 113: // isNil
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		p.SetVarInt(pr, tk.IsNil(v1))

		return ""

	case 121: // test
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0])
		v2 := p.GetVarValue(instrT.Params[1])

		if v1 == v2 {
			tk.Pl("test passed: %#v <-> %#v", v1, v2)
		} else {
			return p.ErrStrf("test failed: %#v <-> %#v", v1, v2)
		}

		return ""

	case 131: // typeOf

		pr := -5

		var v1 interface{}

		if instrT.ParamLen < 1 {
			v1 = p.Peek()
		} else if instrT.ParamLen < 2 {
			v1 = p.GetVarValue(instrT.Params[0])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
		}

		p.SetVarInt(pr, fmt.Sprintf("%T", v1))

		return ""
	case 141: // layer
		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		p.SetVarInt(pr, p.CurrentFuncContextM.Layer)

		return ""
	case 151: // loadCode

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		var codeT string

		codeT = tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT := p.Load(codeT)

		p.SetVarInt(pr, rsT)

		return ""
	case 161: // len
		var lenT int

		var pr int = -5

		var v2 interface{}

		if instrT.ParamLen < 1 {
			v2 = p.Peek()
		} else if instrT.ParamLen < 2 {
			v2 = p.GetVarValue(instrT.Params[0])
		} else {
			pr = instrT.Params[0].Ref
			v2 = p.GetVarValue(instrT.Params[1])
		}

		switch nv := v2.(type) {
		case string:
			lenT = len(nv)
			break
		case []interface{}:
			lenT = len(nv)
			break
		case []string:
			lenT = len(nv)
			break
		case []map[string]string:
			lenT = len(nv)
			break
		case map[string]interface{}:
			lenT = len(nv)
			break
		case map[string]string:
			lenT = len(nv)
			break
		default:
			return p.ErrStrf("无法获取长度的类型：%T", v2)
		}

		p.SetVarInt(pr, lenT)

		return ""
	case 170: // fatalf
		list1T := []interface{}{}

		formatT := ""

		for i, v := range instrT.Params {
			if i == 0 {
				formatT = v.Value.(string)
				continue
			}

			list1T = append(list1T, p.GetVarValue(v))
		}

		fmt.Printf(formatT+"\n", list1T...)

		return "exit"
	case 180: // goto
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		c, ok := v1.(int)

		if ok {
			return c
		}

		s2, ok := v1.(string)

		if !ok {
			return p.ErrStrf("无效的标号：%v", v1)
		}

		if strings.HasPrefix(s2, "+") {
			return p.CodePointerM + tk.ToInt(s2[1:])
		} else if strings.HasPrefix(s2, "-") {
			return p.CodePointerM - tk.ToInt(s2[1:])
		} else {
			labelPointerT, ok := p.LabelsM[p.VarIndexMapM[s2]]

			if ok {
				return labelPointerT
			} else {
				return p.ErrStrf("无效的标号：%v", v1)
			}
		}

		return p.ErrStrf("无效的标号：%v", v1)
	case 199: // exit
		if instrT.ParamLen < 1 {
			return "exit"
		}

		valueT := p.GetVarValue(instrT.Params[0])

		p.SetVar("outG", valueT)

		return "exit"
	case 201: // global
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		// contextT := p.CurrentFuncContextM

		if instrT.ParamLen < 2 {
			p.SetVarIntGlobal(pr, nil)
			// contextT.VarsM[nameT] = ""
			return ""
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		if v1 == "bool" || v1 == "布尔" {
			if instrT.ParamLen > 2 {
				p.SetVarIntGlobal(pr, tk.ToBool(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntGlobal(pr, false)
			}
		} else if v1 == "int" || v1 == "整数" {
			if instrT.ParamLen > 2 {
				p.SetVarIntGlobal(pr, tk.ToInt(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntGlobal(pr, int(0))
			}
		} else if v1 == "byte" || v1 == "字节" {
			if instrT.ParamLen > 2 {
				p.SetVarIntGlobal(pr, tk.ToByte(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntGlobal(pr, byte(0))
			}
		} else if v1 == "rune" || v1 == "如痕" {
			if instrT.ParamLen > 2 {
				p.SetVarIntGlobal(pr, tk.ToRune(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntGlobal(pr, rune(0))
			}
		} else if v1 == "float" || v1 == "小数" {
			if instrT.ParamLen > 2 {
				p.SetVarIntGlobal(pr, tk.ToFloat(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntGlobal(pr, float64(0.0))
			}
		} else if v1 == "str" || v1 == "字符串" {
			if instrT.ParamLen > 2 {
				p.SetVarIntGlobal(pr, tk.ToStr(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntGlobal(pr, "")
			}
		} else if v1 == "list" || v1 == "array" || v1 == "[]" || v1 == "列表" {
			blT := make([]interface{}, 0, instrT.ParamLen-2)

			vs := p.ParamsToList(instrT, v1p+1)

			for _, vvv := range vs {
				nv, ok := vvv.([]interface{})

				if ok {
					for _, vvvj := range nv {
						blT = append(blT, vvvj)
					}
				} else {
					blT = append(blT, vvv)
				}
			}

			p.SetVarIntGlobal(pr, blT)
		} else if v1 == "strList" || v1 == "字符串列表" {
			blT := make([]string, 0, instrT.ParamLen-2)

			vs := p.ParamsToList(instrT, v1p+1)

			for _, vvv := range vs {
				nv, ok := vvv.([]string)

				if ok {
					for _, vvvj := range nv {
						blT = append(blT, vvvj)
					}
				} else {
					blT = append(blT, tk.ToStr(vvv))
				}
			}

			p.SetVarIntGlobal(pr, blT)
		} else if v1 == "byteList" || v1 == "字节列表" {
			blT := make([]byte, 0, instrT.ParamLen-2)

			vs := p.ParamsToList(instrT, v1p+1)

			for _, vvv := range vs {
				nv, ok := vvv.([]byte)

				if ok {
					for _, vvvj := range nv {
						blT = append(blT, vvvj)
					}
				} else {
					blT = append(blT, tk.ToByte(vvv, 0))
				}
			}

			p.SetVarIntGlobal(pr, blT)
		} else if v1 == "runeList" || v1 == "如痕列表" {
			blT := make([]rune, 0, instrT.ParamLen-2)

			vs := p.ParamsToList(instrT, v1p+1)

			for _, vvv := range vs {
				nv, ok := vvv.([]rune)

				if ok {
					for _, vvvj := range nv {
						blT = append(blT, vvvj)
					}
				} else {
					blT = append(blT, tk.ToRune(vvv, 0))
				}
			}

			p.SetVarIntGlobal(pr, blT)
		} else if v1 == "map" || v1 == "映射" {
			p.SetVarIntGlobal(pr, map[string]interface{}{})
		} else if v1 == "strMap" || v1 == "字符串映射" {
			p.SetVarIntGlobal(pr, map[string]string{})
		} else if v1 == "time" || v1 == "时间" || v1 == "time.Time" {
			if instrT.ParamLen > 2 {
				p.SetVarIntGlobal(pr, tk.ToTime(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntGlobal(pr, time.Now())
			}
		} else {
			switch v1 {
			case "gui":
				objT := p.GetVar("guiG")
				p.SetVarIntGlobal(pr, objT)
			case "quickDelegate":
				if instrT.ParamLen < 2 {
					return p.ErrStrf("参数不够")
				}

				v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))

				var deleT tk.QuickDelegate

				// same as fastCall
				deleT = func(strA string) string {
					pointerT := p.CodePointerM

					p.Push(strA)

					tmpPointerT := v2

					for {
						rs := p.RunLine(tmpPointerT)

						nv, ok := rs.(int)

						if ok {
							tmpPointerT = nv
							continue
						}

						nsv, ok := rs.(string)

						if ok {
							if tk.IsErrStr(nsv) {
								// tmpRs := p.Pop()
								p.CodePointerM = pointerT
								return nsv
							}

							if nsv == "exit" { // 不应发生
								tmpRs := p.Pop()
								p.CodePointerM = pointerT
								return tk.ToStr(tmpRs)
							} else if nsv == "fr" {
								break
							}
						}

						tmpPointerT++
					}

					// return pointerT + 1

					tmpRs := p.Pop()
					p.CodePointerM = pointerT
					return tk.ToStr(tmpRs)
				}

				p.SetVarIntGlobal(pr, deleT)
			case "image.Point", "point":
				var p1 image.Point
				if instrT.ParamLen > 3 {
					p1 = image.Point{X: tk.ToInt(p.GetVarValue(instrT.Params[2])), Y: tk.ToInt(p.GetVarValue(instrT.Params[3]))}
					p.SetVarIntGlobal(pr, p1)
				} else {
					p.SetVarIntGlobal(pr, p1)
				}
			default:
				p.SetVarIntGlobal(pr, nil)

			}

		}

		return ""
	case 203: // var
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		// contextT := p.CurrentFuncContextM

		if instrT.ParamLen < 2 {
			p.SetVarIntLocal(pr, nil)
			// contextT.VarsM[nameT] = ""
			return ""
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		if v1 == "bool" || v1 == "布尔" {
			if instrT.ParamLen > 2 {
				p.SetVarIntLocal(pr, tk.ToBool(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntLocal(pr, false)
			}
		} else if v1 == "int" || v1 == "整数" {
			if instrT.ParamLen > 2 {
				p.SetVarIntLocal(pr, tk.ToInt(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntLocal(pr, int(0))
			}
		} else if v1 == "byte" || v1 == "字节" {
			if instrT.ParamLen > 2 {
				p.SetVarIntLocal(pr, tk.ToByte(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntLocal(pr, byte(0))
			}
		} else if v1 == "rune" || v1 == "如痕" {
			if instrT.ParamLen > 2 {
				p.SetVarIntLocal(pr, tk.ToRune(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntLocal(pr, rune(0))
			}
		} else if v1 == "float" || v1 == "小数" {
			if instrT.ParamLen > 2 {
				p.SetVarIntLocal(pr, tk.ToFloat(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntLocal(pr, float64(0.0))
			}
		} else if v1 == "str" || v1 == "字符串" {
			if instrT.ParamLen > 2 {
				p.SetVarIntLocal(pr, tk.ToStr(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntLocal(pr, "")
			}
		} else if v1 == "list" || v1 == "array" || v1 == "[]" || v1 == "列表" {
			blT := make([]interface{}, 0, instrT.ParamLen-2)

			vs := p.ParamsToList(instrT, v1p+1)

			for _, vvv := range vs {
				nv, ok := vvv.([]interface{})

				if ok {
					for _, vvvj := range nv {
						blT = append(blT, vvvj)
					}
				} else {
					blT = append(blT, vvv)
				}
			}

			p.SetVarIntLocal(pr, blT)
		} else if v1 == "strList" || v1 == "字符串列表" {
			blT := make([]string, 0, instrT.ParamLen-2)

			vs := p.ParamsToList(instrT, v1p+1)

			for _, vvv := range vs {
				nv, ok := vvv.([]string)

				if ok {
					for _, vvvj := range nv {
						blT = append(blT, vvvj)
					}
				} else {
					blT = append(blT, tk.ToStr(vvv))
				}
			}

			p.SetVarIntLocal(pr, blT)
		} else if v1 == "byteList" || v1 == "字节列表" {
			blT := make([]byte, 0, instrT.ParamLen-2)

			vs := p.ParamsToList(instrT, v1p+1)

			for _, vvv := range vs {
				nv, ok := vvv.([]byte)

				if ok {
					for _, vvvj := range nv {
						blT = append(blT, vvvj)
					}
				} else {
					blT = append(blT, tk.ToByte(vvv, 0))
				}
			}

			p.SetVarIntLocal(pr, blT)
		} else if v1 == "runeList" || v1 == "如痕列表" {
			blT := make([]rune, 0, instrT.ParamLen-2)

			vs := p.ParamsToList(instrT, v1p+1)

			for _, vvv := range vs {
				nv, ok := vvv.([]rune)

				if ok {
					for _, vvvj := range nv {
						blT = append(blT, vvvj)
					}
				} else {
					blT = append(blT, tk.ToRune(vvv, 0))
				}
			}

			p.SetVarIntLocal(pr, blT)
		} else if v1 == "map" || v1 == "映射" {
			p.SetVarIntLocal(pr, map[string]interface{}{})
		} else if v1 == "strMap" || v1 == "字符串映射" {
			p.SetVarIntLocal(pr, map[string]string{})
		} else if v1 == "time" || v1 == "时间" || v1 == "time.Time" {
			if instrT.ParamLen > 2 {
				p.SetVarIntLocal(pr, tk.ToTime(p.GetVarValue(instrT.Params[2])))
			} else {
				p.SetVarIntLocal(pr, time.Now())
			}
		} else {
			switch v1 {
			case "gui":
				objT := p.GetVar("guiG")
				p.SetVarIntLocal(pr, objT)
			case "quickDelegate":
				if instrT.ParamLen < 2 {
					return p.ErrStrf("参数不够")
				}

				v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))

				var deleT tk.QuickDelegate

				// same as fastCall
				deleT = func(strA string) string {
					pointerT := p.CodePointerM

					p.Push(strA)

					tmpPointerT := v2

					for {
						rs := p.RunLine(tmpPointerT)

						nv, ok := rs.(int)

						if ok {
							tmpPointerT = nv
							continue
						}

						nsv, ok := rs.(string)

						if ok {
							if tk.IsErrStr(nsv) {
								// tmpRs := p.Pop()
								p.CodePointerM = pointerT
								return nsv
							}

							if nsv == "exit" { // 不应发生
								tmpRs := p.Pop()
								p.CodePointerM = pointerT
								return tk.ToStr(tmpRs)
							} else if nsv == "fr" {
								break
							}
						}

						tmpPointerT++
					}

					// return pointerT + 1

					tmpRs := p.Pop()
					p.CodePointerM = pointerT
					return tk.ToStr(tmpRs)
				}

				p.SetVarIntLocal(pr, deleT)
			case "image.Point", "point":
				var p1 image.Point
				if instrT.ParamLen > 3 {
					p1 = image.Point{X: tk.ToInt(p.GetVarValue(instrT.Params[2])), Y: tk.ToInt(p.GetVarValue(instrT.Params[3]))}
					p.SetVarIntLocal(pr, p1)
				} else {
					p.SetVarIntLocal(pr, p1)
				}
			default:
				p.SetVarIntLocal(pr, nil)

			}

		}

		return ""
	case 205: // const
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		p.SetVarInt(pr, ConstMapG[tk.ToStr(p.GetVarValue(instrT.Params[v1p]))])

		return ""
	case 210: // ref
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5

		var v2 interface{}

		if instrT.ParamLen < 2 {
			v2 = p.GetVarRef(instrT.Params[0])
		} else {
			pr = instrT.Params[0].Ref
			v2 = p.GetVarRef(instrT.Params[1])
		}

		p.SetVarInt(pr, v2)

		return ""
	case 211: // refNative
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5

		var v2 interface{}

		if instrT.ParamLen < 2 {
			v2 = p.GetVarRefNative(instrT.Params[0])
		} else {
			pr = instrT.Params[0].Ref
			v2 = p.GetVarRefNative(instrT.Params[1])
		}

		p.SetVarInt(pr, v2)

		return ""
	case 215: // unref
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5

		var v2 interface{}

		if instrT.ParamLen < 1 {
			v2 = p.TmpM
		} else if instrT.ParamLen < 2 {
			v2 = p.GetVarValue(instrT.Params[0])
		} else {
			pr = instrT.Params[0].Ref
			v2 = p.GetVarValue(instrT.Params[1])
		}

		switch nv := v2.(type) {
		case *interface{}:
			p.SetVarInt(pr, *nv)
		case *byte:
			p.SetVarInt(pr, *nv)
		case *int:
			p.SetVarInt(pr, *nv)
		case *uint64:
			p.SetVarInt(pr, *nv)
		case *rune:
			p.SetVarInt(pr, *nv)
		case *bool:
			p.SetVarInt(pr, *nv)
		case *string:
			p.SetVarInt(pr, *nv)
		case *strings.Builder:
			p.SetVarInt(pr, *nv)
		default:
			return p.ErrStrf("无法处理的类型：%T", v2)
		}

		return ""
	case 218: // assignRef
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		// nameT := instrT.Params[0].Ref
		p1a := p.GetVarValue(instrT.Params[0])

		// p1, ok := p1a.(*interface{})

		var p1 *interface{}
		v1p := 1

		switch nv := p1a.(type) {
		case *interface{}:
			p1 = nv
			break
		case *byte:
			*nv = tk.ToByte(p.GetVarValue(instrT.Params[v1p]))
			return ""
		case *rune:
			*nv = tk.ToRune(p.GetVarValue(instrT.Params[v1p]))
			return ""
		case *int:
			*nv = tk.ToInt(p.GetVarValue(instrT.Params[v1p]))
			return ""
		default:
			return p.ErrStrf("无法处理的类型：%T", p1a)
		}

		if instrT.ParamLen > 2 {
			valueTypeT := instrT.Params[1].Value
			valueT := p.GetVarValue(instrT.Params[2])

			if valueTypeT == "bool" || valueTypeT == "布尔" {
				*p1 = tk.ToBool(valueT)
			} else if valueTypeT == "int" || valueTypeT == "整数" {
				*p1 = tk.ToInt(valueT)
			} else if valueTypeT == "byte" || valueTypeT == "字节" {
				*p1 = tk.ToByte(valueT)
			} else if valueTypeT == "rune" || valueTypeT == "如痕" {
				*p1 = tk.ToRune(valueT)
			} else if valueTypeT == "float" || valueTypeT == "小数" {
				*p1 = tk.ToFloat(valueT)
			} else if valueTypeT == "str" || valueTypeT == "字符串" {
				*p1 = tk.ToStr(valueT)
			} else if valueTypeT == "list" || valueT == "array" || valueT == "[]" || valueTypeT == "列表" {
				*p1 = valueT.([]interface{})
			} else if valueTypeT == "strList" || valueTypeT == "字符串列表" {
				*p1 = valueT.([]string)
			} else if valueTypeT == "byteList" || valueTypeT == "字节列表" {
				*p1 = valueT.([]byte)
			} else if valueTypeT == "runeList" || valueTypeT == "如痕列表" {
				*p1 = valueT.([]rune)
			} else if valueTypeT == "map" || valueTypeT == "映射" {
				*p1 = valueT.(map[string]interface{})
			} else if valueTypeT == "strMap" || valueTypeT == "字符串映射" {
				*p1 = valueT.(map[string]string)
			} else if valueTypeT == "time" || valueTypeT == "时间" {
				*p1 = valueT.(time.Time)
			} else {
				*p1 = valueT
			}

			return ""
		}

		valueT := p.GetVarValue(instrT.Params[1])

		*p1 = valueT

		// (*(p.CurrentVarsM))[nameT] = valueT

		return ""
	case 220: // push
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		if instrT.ParamLen > 1 {
			v2 := p.GetVarValue(instrT.Params[1])

			v1 := tk.ToStr(p.GetVarValue(instrT.Params[0]))

			if v1 == "int" || v1 == "整数" {
				p.Push(tk.ToInt(v2))
			} else if v1 == "byte" || v1 == "字节" {
				p.Push(tk.ToByte(v2))
			} else if v1 == "rune" || v1 == "如痕" {
				p.Push(tk.ToRune(v2))
			} else if v1 == "float" || v1 == "小数" {
				p.Push(tk.ToFloat(v2))
			} else if v1 == "bool" || v1 == "布尔" {
				p.Push(tk.ToBool(v2))
			} else if v1 == "str" || v1 == "字符串" {
				p.Push(tk.ToStr(v2))
			} else {
				p.Push(v2)
			}

			return ""
		}

		v1 := p.GetVarValue(instrT.Params[0])

		p.Push(v1)

		return ""
	case 222: // peek
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		// if pr == -5 {
		// 	p.Push(p.Peek())
		// 	return ""
		// }

		// if p1 < 0 {
		// 	return p.ErrStrf("invalid var name")
		// }

		errT := p.SetVarInt(pr, p.Peek())

		if errT != nil {
			return p.ErrStrf("%v", errT)
		}

		// (*(p.CurrentVarsM))[p1] = p.Peek()

		return ""
	// case 223: // peek$
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	pr := instrT.Params[0].Ref

	// 	// if p1 < 0 {
	// 	// 	return p.ErrStrf("invalid var name")
	// 	// }

	// 	p.SetVarInt(pr, p.Peek())
	// 	// (*(p.CurrentVarsM))[p1] = p.Peek()

	// 	return ""
	case 224: // pop
		pr := -5
		if instrT.ParamLen < 1 {
			p.SetVarInt(pr, p.Pop())
			return ""
		}

		pr = instrT.Params[0].Ref

		// if p1 < 0 {
		// 	return p.ErrStrf("invalid var name")
		// }

		p.SetVarInt(pr, p.Pop())
		// (*(p.CurrentVarsM))[p1] = p.Pop()

		return ""
	// case 226: // peek*
	// 	v1 := instrT.Params[0].Value.(int)

	// 	p.CurrentFuncContextM.RegsM.IntsM[v1] = p.Peek().(int)

	// 	return ""
	// case 227: // pop*
	// 	v1 := instrT.Params[0].Value.(int)

	// 	p.CurrentFuncContextM.RegsM.IntsM[v1] = p.Pop().(int)

	// 	return ""
	// case 231: // pushInt
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	v1 := p.GetVarValue(instrT.Params[0])

	// 	if tk.IsError(v1) {
	// 		return p.ErrStrf("参数错误")
	// 	}

	// 	p.Push(tk.ToInt(v1))

	// 	return ""
	// case 232: // pushInt$
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}
	// 	// tk.Plv(p.GetVars())

	// 	v1 := p.GetVarValue(instrT.Params[0])

	// 	if tk.IsError(v1) {
	// 		// p.Debug()
	// 		// tk.Plv(instrT)
	// 		return p.ErrStrf("参数错误: %v", v1)
	// 	}

	// 	cT, ok := v1.(int)
	// 	if ok {
	// 		p.Push(cT)
	// 		return ""
	// 	}

	// 	sT, ok := v1.(string)
	// 	if ok {
	// 		c1T, errT := tk.StrToIntQuick(sT)

	// 		if errT != nil {
	// 			return p.ErrStrf("convert value to int failed: %v", errT)
	// 		}

	// 		p.Push(c1T)

	// 		return ""
	// 	}

	// 	return p.ErrStrf("无效的数据格式")
	// case 233: // pushInt#
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	v1 := instrT.Params[0].Value.(int)

	// 	// c1T, errT := tk.StrToIntQuick(v1)

	// 	// if errT != nil {
	// 	// 	return p.ErrStrf("convert value to int failed: %v", errT)
	// 	// }

	// 	p.Push(v1)

	// 	return ""
	// case 234: // pushInt*
	// 	v1 := instrT.Params[0].Value.(int)

	// 	p.Push(p.CurrentFuncContextM.RegsM.IntsM[v1])

	// 	return ""
	// case 290: // pushLocal
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	v1 := p.GetVarValue(instrT.Params[0])

	// 	if tk.IsError(v1) {
	// 		return p.ErrStrf("invalid param")
	// 	}

	// 	p.Push(v1)

	// 	return ""
	case 240: // clearStack
		// p.StackM = make([]interface{}, 0, 10)
		p.StackPointerM = 0

	// case 312: // regInt#  from value
	// 	v1 := instrT.Params[0].Value.(int)
	// 	v2 := instrT.Params[1].Value.(int)

	// 	p.CurrentFuncContextM.RegsM.IntsM[v1] = v2

	// 	return ""

	case 301: // getSharedMapItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref // -5
		v1p := 1

		// if instrT.ParamLen > 2 {
		// 	pr = instrT.Params[0].Ref
		// 	v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		// v1 := p.GetVarValue(instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// var vT interface{}
		// tk.Pln(pr, v1, v2)

		var defaultT interface{} = Undefined

		if instrT.ParamLen > 2 {
			defaultT = p.GetVarValue(instrT.Params[2])
		}

		p.SetVarInt(pr, p.SharedMapM.Get(v1, defaultT))

		return ""

	case 302: // getSharedMapSize

		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		p.SetVarInt(pr, p.SharedMapM.Size())

		return ""

	case 303: // tryGetSharedMapItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref // -5
		v1p := 1

		// if instrT.ParamLen > 2 {
		// 	pr = instrT.Params[0].Ref
		// 	v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		// v1 := p.GetVarValue(instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// var vT interface{}
		// tk.Pln(pr, v1, v2)

		var defaultT interface{} = Undefined

		if instrT.ParamLen > 2 {
			defaultT = p.GetVarValue(instrT.Params[2])
		}

		p.SetVarInt(pr, p.SharedMapM.TryGet(v1, defaultT))

		return ""

	case 304: // tryGetSharedMapSize
		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		vr := p.SharedMapM.TrySize()

		if vr < 0 {
			p.SetVarInt(pr, fmt.Errorf("获取大小失败"))
		} else {
			p.SetVarInt(pr, vr)
		}

		return ""

	case 311: // setSharedMapItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		// pr := -5
		v1p := 0

		// if instrT.ParamLen > 2 {
		// 	// pr = instrT.Params[0].Ref
		// 	v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		// v1 := p.GetVarValue(instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := p.GetVarValue(instrT.Params[v1p+1])

		p.SharedMapM.Set(v1, v2)
		// p.SetVarInt(pr, true)

		return ""

	case 313: // trySetSharedMapItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
			// p.SetVarInt(instrT.Params[2].Ref, vT)
		}

		// v1 := p.GetVarValue(instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := p.GetVarValue(instrT.Params[v1p+1])

		p.SetVarInt(pr, p.SharedMapM.TrySet(v1, v2))

		return ""

	case 321: // deleteSharedMapItem
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		// pr := -5
		v1p := 0

		// if instrT.ParamLen > 1 {
		// 	// pr = instrT.Params[0].Ref
		// 	v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		// v1 := p.GetVarValue(instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SharedMapM.Delete(v1)
		// p.SetVarInt(pr, true)

		return ""

	case 323: // tryDeleteSharedMapItem
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
			// p.SetVarInt(instrT.Params[2].Ref, vT)
		}

		// v1 := p.GetVarValue(instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, p.SharedMapM.TryDelete(v1))

		return ""

	case 331: // clearSharedMapItem
		// if instrT.ParamLen < 1 {
		// 	return p.ErrStrf("参数不够")
		// }

		// pr := -5
		// // v1p := 0

		// if instrT.ParamLen > 1 {
		// 	pr = instrT.Params[0].Ref
		// 	// v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		// v1 := p.GetVarValue(instrT.Params[v1p])

		// v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SharedMapM.Clear()
		// p.SetVarInt(pr, true)

		return ""

	case 333: // tryClearSharedMap
		// if instrT.ParamLen < 1 {
		// 	return p.ErrStrf("参数不够")
		// }

		pr := -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			// v1p = 1
			// p.SetVarInt(instrT.Params[2].Ref, vT)
		}

		// v1 := p.GetVarValue(instrT.Params[v1p])

		// v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, p.SharedMapM.TryClear())

		return ""

	case 341: // lockSharedMap
		p.SharedMapM.Lock()

		return ""

	case 342: // tryLockSharedMap
		p.SharedMapM.TryLock()

		return ""

	case 343: // unlockSharedMap
		p.SharedMapM.Unlock()

		return ""

	case 346: // readLockSharedMap
		p.SharedMapM.RLock()

		return ""

	case 347: // tryReadLockSharedMap
		p.SharedMapM.TryRLock()

		return ""

	case 348: // readUnlockSharedMap
		p.SharedMapM.RUnlock()

		return ""

	case 351: // quickClearSharedMap
		p.SharedMapM.QuickClear()

		return ""

	case 353: // quickGetSharedMapItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref // -5
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		var defaultT interface{} = Undefined

		if instrT.ParamLen > 2 {
			defaultT = p.GetVarValue(instrT.Params[2])
		}

		p.SetVarInt(pr, p.SharedMapM.QuickGet(v1, defaultT))

		return ""

	case 354: // quickGetSharedMap
		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		p.SetVarInt(pr, p.SharedMapM)

		return ""

	case 355: // quickSetSharedMapItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1p := 0

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := p.GetVarValue(instrT.Params[v1p+1])

		p.SharedMapM.QuickSet(v1, v2)

		return ""

	case 357: // quickDeleteSharedMapItem
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		v1p := 0

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SharedMapM.QuickDelete(v1)

		return ""
	case 359: // quickSizeSharedMap

		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		p.SetVarInt(pr, p.SharedMapM.QuickSize())

		return ""

	case 401: // assign/=/赋值
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		if instrT.ParamLen > 2 {
			valueTypeT := instrT.Params[1].Value
			valueT := p.GetVarValue(instrT.Params[2])

			if valueTypeT == "bool" || valueTypeT == "布尔" {
				p.SetVarInt(pr, tk.ToBool(valueT))
			} else if valueTypeT == "int" || valueTypeT == "整数" {
				p.SetVarInt(pr, tk.ToInt(valueT))
			} else if valueTypeT == "byte" || valueTypeT == "字节" {
				p.SetVarInt(pr, tk.ToByte(valueT))
			} else if valueTypeT == "rune" || valueTypeT == "如痕" {
				p.SetVarInt(pr, tk.ToRune(valueT))
			} else if valueTypeT == "float" || valueTypeT == "小数" {
				p.SetVarInt(pr, tk.ToFloat(valueT))
			} else if valueTypeT == "str" || valueTypeT == "字符串" {
				p.SetVarInt(pr, tk.ToStr(valueT))
			} else if valueTypeT == "list" || valueT == "array" || valueT == "[]" || valueTypeT == "列表" {
				p.SetVarInt(pr, valueT.([]interface{}))
			} else if valueTypeT == "strList" || valueTypeT == "字符串列表" {
				p.SetVarInt(pr, valueT.([]string))
			} else if valueTypeT == "byteList" || valueTypeT == "字节列表" {
				p.SetVarInt(pr, valueT.([]byte))
			} else if valueTypeT == "runeList" || valueTypeT == "如痕列表" {
				p.SetVarInt(pr, valueT.([]rune))
			} else if valueTypeT == "map" || valueTypeT == "映射" {
				p.SetVarInt(pr, valueT.(map[string]interface{}))
			} else if valueTypeT == "strMap" || valueTypeT == "字符串映射" {
				p.SetVarInt(pr, valueT.(map[string]string))
			} else if valueTypeT == "time" || valueTypeT == "时间" {
				p.SetVarInt(pr, valueT.(map[string]string))
			} else {
				p.SetVarInt(pr, valueT)
			}

			return ""
		}

		valueT := p.GetVarValue(instrT.Params[1])

		p.SetVarInt(pr, valueT)

		// (*(p.CurrentVarsM))[nameT] = valueT

		return ""
	case 402: // assign$
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		// if nameT < 0 {
		// 	return p.ErrStrf("invalid var name")
		// }

		p.SetVarInt(pr, p.Pop())
		// (*(p.CurrentVarsM))[nameT] = p.Pop()

		return ""
	// case 410: // assignInt
	// 	if instrT.ParamLen < 2 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	pr := instrT.Params[0].Ref

	// 	// if nameT < 0 {
	// 	// 	return p.ErrStrf("invalid var name")
	// 	// }

	// 	valueT := p.GetVarValue(instrT.Params[1])

	// 	p.SetVarInt(pr, tk.ToInt(valueT))

	// 	return ""
	// case 411: // assignI
	// 	if instrT.ParamLen < 2 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	p1 := instrT.Params[0].Ref

	// 	// if p1 <= 0 {
	// 	// 	return p.ErrStrf("invalid var name")
	// 	// }

	// 	v2 := instrT.Params[1].Value.(string)

	// 	c2T, errT := tk.StrToIntQuick(v2)

	// 	if errT != nil {
	// 		return p.ErrStrf("convert value to int failed: %v", errT)
	// 	}

	// 	p.SetVarInt(p1, c2T)

	// 	return ""
	case 491: // assignGlobal
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		// if nameT < 0 {
		// 	return p.ErrStrf("invalid var name")
		// }

		valueT := p.GetVarValue(instrT.Params[1])

		p.SetVarIntGlobal(pr, valueT)

		// p.VarsM[nameT] = valueT

		return ""
	case 492: // assignFromGlobal
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		// if p1 <= 0 {
		// 	return p.ErrStrf("invalid var name")
		// }

		valueT := p.GetVarValueGlobal(instrT.Params[1])

		p.SetVarInt(pr, valueT)

		return ""
	case 493: // assignLocal/局部赋值
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		if instrT.ParamLen > 2 {
			valueTypeT := instrT.Params[1].Value
			valueT := p.GetVarValue(instrT.Params[2])

			if valueTypeT == "bool" || valueTypeT == "布尔" {
				p.SetVarIntLocal(pr, tk.ToBool(valueT))
			} else if valueTypeT == "int" || valueTypeT == "整数" {
				p.SetVarIntLocal(pr, tk.ToInt(valueT))
			} else if valueTypeT == "byte" || valueTypeT == "字节" {
				p.SetVarIntLocal(pr, tk.ToByte(valueT))
			} else if valueTypeT == "rune" || valueTypeT == "如痕" {
				p.SetVarIntLocal(pr, tk.ToRune(valueT))
			} else if valueTypeT == "float" || valueTypeT == "小数" {
				p.SetVarIntLocal(pr, tk.ToFloat(valueT))
			} else if valueTypeT == "str" || valueTypeT == "字符串" {
				p.SetVarIntLocal(pr, tk.ToStr(valueT))
			} else if valueTypeT == "list" || valueTypeT == "列表" {
				p.SetVarIntLocal(pr, valueT.([]interface{}))
			} else if valueTypeT == "strList" || valueTypeT == "字符串列表" {
				p.SetVarIntLocal(pr, valueT.([]string))
			} else if valueTypeT == "byteList" || valueTypeT == "字节列表" {
				p.SetVarIntLocal(pr, valueT.([]byte))
			} else if valueTypeT == "runeList" || valueTypeT == "如痕列表" {
				p.SetVarIntLocal(pr, valueT.([]string))
			} else if valueTypeT == "map" || valueTypeT == "映射" {
				p.SetVarIntLocal(pr, valueT.(map[string]interface{}))
			} else if valueTypeT == "strMap" || valueTypeT == "字符串映射" {
				p.SetVarIntLocal(pr, valueT.(map[string]string))
			} else if valueTypeT == "time" || valueTypeT == "时间" {
				p.SetVarIntLocal(pr, valueT.(time.Time))
			} else {
				p.SetVarIntLocal(pr, valueT)
			}

			return ""
		}

		valueT := p.GetVarValue(instrT.Params[1])

		p.SetVarIntLocal(pr, valueT)

		// (*(p.CurrentVarsM))[nameT] = valueT

		return ""
	case 610: // if
		// tk.Plv(instrT)
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var condT bool
		var v2 interface{}
		var ok0 bool

		var elseLabelIntT int = -1

		if instrT.ParamLen > 2 {
			elseLabelT := p.GetVarValue(instrT.Params[2])

			s2, sok := elseLabelT.(int)

			if sok {
				elseLabelIntT = s2
			} else {
				st2, stok := elseLabelT.(string)

				if !stok {
					return p.ErrStrf("无效的标号：%v", elseLabelT)
				}

				if strings.HasPrefix(st2, "+") {
					elseLabelIntT = p.CodePointerM + tk.ToInt(st2[1:])
				} else if strings.HasPrefix(st2, "-") {
					elseLabelIntT = p.CodePointerM - tk.ToInt(st2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[st2]]

					if ok {
						elseLabelIntT = labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", elseLabelT)
					}
				}
			}
		}
		// tk.Plv(instrT)
		if instrT.ParamLen < 2 {
			tmpv := p.Pop()
			condT, ok0 = tmpv.(bool)
			v2 = p.GetVarValue(instrT.Params[0])

			if !ok0 {
				return p.ErrStrf("无效的参数：%#v", tmpv)
			}

		} else {
			tmpv := p.GetVarValue(instrT.Params[0])
			condT, ok0 = tmpv.(bool)
			v2 = p.GetVarValue(instrT.Params[1])

			if !ok0 {
				var tmps string
				tmps, ok0 = tmpv.(string)

				if ok0 {
					tmprs := p.QuickEval(tmps)

					condT, ok0 = tmprs.(bool)
				}
			}

			if !ok0 {
				return p.ErrStrf("无效的参数：%#v", tmpv)
			}
		}

		s2, sok := v2.(string)

		if !sok {
			if condT {
				c2, cok := v2.(int)
				if cok {
					// tk.Pln("c2", c2)
					return c2
				} else {
					return p.ErrStrf("无效的标号：%v", v2)
				}
			}
		} else {
			if condT {
				if strings.HasPrefix(s2, "+") {
					return p.CodePointerM + tk.ToInt(s2[1:])
				} else if strings.HasPrefix(s2, "-") {
					return p.CodePointerM - tk.ToInt(s2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[s2]]
					// tk.Pln("labelPointerT", labelPointerT, ok)

					if ok {
						return labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", v2)
					}
				}
			}
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 611: // ifNot
		// tk.Plv(instrT)
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var condT bool
		var v2 interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 2 {
			elseLabelT := p.GetVarValue(instrT.Params[2])

			s2, sok := elseLabelT.(int)

			if sok {
				elseLabelIntT = s2
			} else {
				st2, stok := elseLabelT.(string)

				if !stok {
					return p.ErrStrf("无效的标号：%v", elseLabelT)
				}

				if strings.HasPrefix(st2, "+") {
					elseLabelIntT = p.CodePointerM + tk.ToInt(st2[1:])
				} else if strings.HasPrefix(st2, "-") {
					elseLabelIntT = p.CodePointerM - tk.ToInt(st2[1:])
				} else {

					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[st2]]

					if ok {
						elseLabelIntT = labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", elseLabelT)
					}
				}
			}
		}

		if instrT.ParamLen < 2 {
			condT = p.Pop().(bool)
			v2 = p.GetVarValue(instrT.Params[0])

		} else {
			condT = p.GetVarValue(instrT.Params[0]).(bool)
			v2 = p.GetVarValue(instrT.Params[1])
		}

		s2, sok := v2.(string)

		if !sok {
			if !condT {
				c2, cok := v2.(int)
				if cok {
					return c2
				} else {
					return p.ErrStrf("无效的标号：%v", v2)
				}
			}
		} else {
			if !condT {
				if strings.HasPrefix(s2, "+") {
					return p.CodePointerM + tk.ToInt(s2[1:])
				} else if strings.HasPrefix(s2, "-") {
					return p.CodePointerM - tk.ToInt(s2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[s2]]

					if ok {
						return labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", v2)
					}
				}
			}
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	// case 618: // if$
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	condT := p.Pop().(bool)

	// 	if condT {
	// 		return p.GetVarValue(instrT.Params[0]).(int)
	// 	}

	// 	return ""

	// case 619: // if*
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	condT := p.CurrentFuncContextM.RegsM.CondsM[0]

	// 	if condT {
	// 		return p.GetVarValue(instrT.Params[0]).(int)
	// 	}

	// 	return ""

	// case 621: // ifNot$
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	condT := p.Pop().(bool)

	// 	if !condT {
	// 		return p.GetVarValue(instrT.Params[0]).(int)
	// 	}

	// 	return ""

	case 631: // ifEval
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var condT bool
		var v2 interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 2 {
			elseLabelT := p.GetVarValue(instrT.Params[2])

			s2, sok := elseLabelT.(int)

			if sok {
				elseLabelIntT = s2
			} else {
				st2, stok := elseLabelT.(string)

				if !stok {
					return p.ErrStrf("无效的标号：%v", elseLabelT)
				}

				if strings.HasPrefix(st2, "+") {
					elseLabelIntT = p.CodePointerM + tk.ToInt(st2[1:])
				} else if strings.HasPrefix(st2, "-") {
					elseLabelIntT = p.CodePointerM - tk.ToInt(st2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[st2]]

					if ok {
						elseLabelIntT = labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", elseLabelT)
					}
				}
			}
		}

		condExprT := p.GetVarValue(instrT.Params[0]).(string)

		condT = p.EvalExpression(condExprT).(bool)

		v2 = p.GetVarValue(instrT.Params[1])

		s2, sok := v2.(string)

		if !sok {
			if condT {
				c2, cok := v2.(int)
				if cok {
					return c2
				} else {
					return p.ErrStrf("无效的标号：%v", v2)
				}
			}
		} else {
			if condT {
				if strings.HasPrefix(s2, "+") {
					return p.CodePointerM + tk.ToInt(s2[1:])
				} else if strings.HasPrefix(s2, "-") {
					return p.CodePointerM - tk.ToInt(s2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[s2]]

					if ok {
						return labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", v2)
					}
				}
			}
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 641: // ifEmpty
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var condT bool
		var v2 interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 2 {
			elseLabelT := p.GetVarValue(instrT.Params[2])

			s2, sok := elseLabelT.(int)

			if sok {
				elseLabelIntT = s2
			} else {
				st2, stok := elseLabelT.(string)

				if !stok {
					return p.ErrStrf("无效的标号：%v", elseLabelT)
				}

				if strings.HasPrefix(st2, "+") {
					elseLabelIntT = p.CodePointerM + tk.ToInt(st2[1:])
				} else if strings.HasPrefix(st2, "-") {
					elseLabelIntT = p.CodePointerM - tk.ToInt(st2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[st2]]

					if ok {
						elseLabelIntT = labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", elseLabelT)
					}
				}
			}
		}

		v1 := p.GetVarValue(instrT.Params[0])

		if v1 == nil {
			condT = true
		} else if v1 == Undefined {
			condT = true
		} else {
			switch nv := v1.(type) {
			case bool:
				condT = (nv == false)
			case string:
				condT = (nv == "")
			case byte:
				condT = (nv <= 0)
			case int:
				condT = (nv <= 0)
			case rune:
				condT = (nv <= 0)
			case int64:
				condT = (nv <= 0)
			case float64:
				condT = (nv <= 0)
			case []byte:
				condT = (len(nv) < 1)
			case []int:
				condT = (len(nv) < 1)
			case []rune:
				condT = (len(nv) < 1)
			case []int64:
				condT = (len(nv) < 1)
			case []float64:
				condT = (len(nv) < 1)
			case []string:
				condT = (len(nv) < 1)
			case []interface{}:
				condT = (len(nv) < 1)
			case map[string]string:
				condT = (len(nv) < 1)
			case map[string]interface{}:
				condT = (len(nv) < 1)
			default:
				condT = false
			}
		}

		v2 = p.GetVarValue(instrT.Params[1])

		s2, sok := v2.(string)

		if !sok {
			if condT {
				c2, cok := v2.(int)
				if cok {
					return c2
				} else {
					return p.ErrStrf("无效的标号：%v", v2)
				}
			}
		} else {
			if condT {
				if strings.HasPrefix(s2, "+") {
					return p.CodePointerM + tk.ToInt(s2[1:])
				} else if strings.HasPrefix(s2, "-") {
					return p.CodePointerM - tk.ToInt(s2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[s2]]

					if ok {
						return labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", v2)
					}
				}
			}
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 643: // ifEqual
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		var condT bool
		var v3 interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 3 {
			elseLabelT := p.GetVarValue(instrT.Params[3])

			s2, sok := elseLabelT.(int)

			if sok {
				elseLabelIntT = s2
			} else {
				st2, stok := elseLabelT.(string)

				if !stok {
					return p.ErrStrf("无效的标号：%v", elseLabelT)
				}

				if strings.HasPrefix(st2, "+") {
					elseLabelIntT = p.CodePointerM + tk.ToInt(st2[1:])
				} else if strings.HasPrefix(st2, "-") {
					elseLabelIntT = p.CodePointerM - tk.ToInt(st2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[st2]]

					if ok {
						elseLabelIntT = labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", elseLabelT)
					}
				}
			}
		}

		v1 := p.GetVarValue(instrT.Params[0])
		v2 := p.GetVarValue(instrT.Params[1])

		if v1 == v2 {
			condT = true
		} else {
			condT = false

		}

		v3 = p.GetVarValue(instrT.Params[2])

		s2, sok := v3.(string)

		if !sok {
			if condT {
				c2, cok := v3.(int)
				if cok {
					return c2
				} else {
					return p.ErrStrf("无效的标号：%v", v3)
				}
			}
		} else {
			if condT {
				if strings.HasPrefix(s2, "+") {
					return p.CodePointerM + tk.ToInt(s2[1:])
				} else if strings.HasPrefix(s2, "-") {
					return p.CodePointerM - tk.ToInt(s2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[s2]]

					if ok {
						return labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", v3)
					}
				}
			}
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 644: // ifNotEqual
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		var condT bool
		var v3 interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 3 {
			elseLabelT := p.GetVarValue(instrT.Params[3])

			s2, sok := elseLabelT.(int)

			if sok {
				elseLabelIntT = s2
			} else {
				st2, stok := elseLabelT.(string)

				if !stok {
					return p.ErrStrf("无效的标号：%v", elseLabelT)
				}

				if strings.HasPrefix(st2, "+") {
					elseLabelIntT = p.CodePointerM + tk.ToInt(st2[1:])
				} else if strings.HasPrefix(st2, "-") {
					elseLabelIntT = p.CodePointerM - tk.ToInt(st2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[st2]]

					if ok {
						elseLabelIntT = labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", elseLabelT)
					}
				}
			}
		}

		v1 := p.GetVarValue(instrT.Params[0])
		v2 := p.GetVarValue(instrT.Params[1])

		if v1 == v2 {
			condT = false
		} else {
			condT = true

		}

		v3 = p.GetVarValue(instrT.Params[2])

		s2, sok := v3.(string)

		if !sok {
			if condT {
				c2, cok := v3.(int)
				if cok {
					return c2
				} else {
					return p.ErrStrf("无效的标号：%v", v3)
				}
			}
		} else {
			if condT {
				if strings.HasPrefix(s2, "+") {
					return p.CodePointerM + tk.ToInt(s2[1:])
				} else if strings.HasPrefix(s2, "-") {
					return p.CodePointerM - tk.ToInt(s2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[s2]]

					if ok {
						return labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", v3)
					}
				}
			}
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 651: // ifErrX
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var condT bool
		var v2 interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 2 {
			elseLabelT := p.GetVarValue(instrT.Params[2])

			s2, sok := elseLabelT.(int)

			if sok {
				elseLabelIntT = s2
			} else {
				st2, stok := elseLabelT.(string)

				if !stok {
					return p.ErrStrf("无效的标号：%v", elseLabelT)
				}

				if strings.HasPrefix(st2, "+") {
					elseLabelIntT = p.CodePointerM + tk.ToInt(st2[1:])
				} else if strings.HasPrefix(st2, "-") {
					elseLabelIntT = p.CodePointerM - tk.ToInt(st2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[st2]]

					if ok {
						elseLabelIntT = labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", elseLabelT)
					}
				}
			}
		}

		v1 := p.GetVarValue(instrT.Params[0])

		if v1 == nil {
			condT = false
		} else if v1 == Undefined {
			condT = false
		} else {
			switch nv := v1.(type) {
			case error:
				if nv == nil {
					condT = false
				} else {
					condT = true
				}
			case string:
				condT = tk.IsErrStr(nv)
			default:
				condT = false
			}
		}

		v2 = p.GetVarValue(instrT.Params[1])

		s2, sok := v2.(string)

		if !sok {
			if condT {
				c2, cok := v2.(int)
				if cok {
					return c2
				} else {
					return p.ErrStrf("无效的标号：%v", v2)
				}
			}
		} else {
			if condT {
				if strings.HasPrefix(s2, "+") {
					return p.CodePointerM + tk.ToInt(s2[1:])
				} else if strings.HasPrefix(s2, "-") {
					return p.CodePointerM - tk.ToInt(s2[1:])
				} else {
					labelPointerT, ok := p.LabelsM[p.VarIndexMapM[s2]]

					if ok {
						return labelPointerT
					} else {
						return p.ErrStrf("无效的标号：%v", v2)
					}
				}
			}
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 701: // ==
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		switch nv := v1.(type) {
		case time.Time:
			rsT := tk.ToTime(v2)
			if tk.IsError(rsT) {
				return p.ErrStrf("时间转换失败：%v", rsT)
			}

			resultT := nv.Equal(rsT.(time.Time))
			p.SetVarInt(pr, resultT)
			return ""

		}

		p.SetVarInt(pr, (v1 == v2))

		return ""
	case 702: // !=
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		switch nv := v1.(type) {
		case time.Time:
			rsT := tk.ToTime(v2)
			if tk.IsError(rsT) {
				return p.ErrStrf("时间转换失败：%v", rsT)
			}

			resultT := !nv.Equal(rsT.(time.Time))
			p.SetVarInt(pr, resultT)
			return ""

		}

		p.SetVarInt(pr, (v1 != v2))

		return ""
	case 703: // <
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		var v3 bool

		switch nv := v1.(type) {
		case int:
			v3 = nv < v2.(int)
		case byte:
			v3 = nv < v2.(byte)
		case rune:
			v3 = nv < v2.(rune)
		case float64:
			v3 = nv < v2.(float64)
		case string:
			v3 = nv < v2.(string)
		default:
			return p.ErrStrf("数据类型不匹配")
		}

		switch nv := v1.(type) {
		case time.Time:
			rsT := tk.ToTime(v2)
			if tk.IsError(rsT) {
				return p.ErrStrf("时间转换失败：%v", rsT)
			}

			resultT := nv.Before(rsT.(time.Time))
			p.SetVarInt(pr, resultT)
			return ""

		}
		p.SetVarInt(pr, v3)

		return ""

	case 704: // >
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		var v3 bool

		switch nv := v1.(type) {
		case int:
			v3 = nv > v2.(int)
		case byte:
			v3 = nv > v2.(byte)
		case rune:
			v3 = nv > v2.(rune)
		case float64:
			v3 = nv > v2.(float64)
		case string:
			v3 = nv > v2.(string)
		default:
			return p.ErrStrf("数据类型不匹配")
		}

		switch nv := v1.(type) {
		case time.Time:
			rsT := tk.ToTime(v2)
			if tk.IsError(rsT) {
				return p.ErrStrf("时间转换失败：%v", rsT)
			}

			resultT := nv.After(rsT.(time.Time))
			p.SetVarInt(pr, resultT)
			return ""

		}
		p.SetVarInt(pr, v3)

		return ""

	case 705: // <=
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		var v3 bool

		switch nv := v1.(type) {
		case int:
			v3 = nv <= v2.(int)
		case byte:
			v3 = nv <= v2.(byte)
		case rune:
			v3 = nv <= v2.(rune)
		case float64:
			v3 = nv <= v2.(float64)
		case string:
			v3 = nv <= v2.(string)
		default:
			return p.ErrStrf("数据类型不匹配")
		}

		switch nv := v1.(type) {
		case time.Time:
			rsT := tk.ToTime(v2)
			if tk.IsError(rsT) {
				return p.ErrStrf("时间转换失败：%v", rsT)
			}

			resultT := nv.Before(rsT.(time.Time)) || nv.Equal(rsT.(time.Time))
			p.SetVarInt(pr, resultT)
			return ""

		}
		p.SetVarInt(pr, v3)

		return ""

	case 706: // >=
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		var v3 bool

		switch nv := v1.(type) {
		case int:
			v3 = nv >= v2.(int)
		case byte:
			v3 = nv >= v2.(byte)
		case rune:
			v3 = nv >= v2.(rune)
		case float64:
			v3 = nv >= v2.(float64)
		case string:
			v3 = nv >= v2.(string)
		default:
			return p.ErrStrf("数据类型不匹配")
		}

		switch nv := v1.(type) {
		case time.Time:
			rsT := tk.ToTime(v2)
			if tk.IsError(rsT) {
				return p.ErrStrf("时间转换失败：%v", rsT)
			}

			resultT := nv.After(rsT.(time.Time)) || nv.Equal(rsT.(time.Time))
			p.SetVarInt(pr, resultT)
			return ""

		}

		p.SetVarInt(pr, v3)

		return ""

	// case 710: // >i
	// 	if instrT.ParamLen < 2 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	s1 := p.GetVarValue(instrT.Params[0]).(int)

	// 	s2, errT := tk.StrToIntQuick(p.GetVarValue(instrT.Params[1]).(string))

	// 	if errT != nil {
	// 		return p.ErrStrf("failed to convert to int: %v", errT)
	// 	}

	// 	p.Push(s1 > s2)

	// 	return ""

	// case 720: // <i
	// 	if instrT.ParamLen < 2 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	var errT error

	// 	p1 := p.GetVarValue(instrT.Params[0])

	// 	c1, ok := p1.(int)

	// 	if !ok {
	// 		s1, ok := p1.(string)

	// 		if ok {
	// 			c1, errT = tk.StrToIntQuick(s1)

	// 			if errT != nil {
	// 				return p.ErrStrf("failed to convert to int: %v", errT)
	// 			}
	// 		} else {
	// 			c1 = tk.ToInt(p1)
	// 		}
	// 	}

	// 	p2 := p.GetVarValue(instrT.Params[1])

	// 	c2, ok := p2.(int)

	// 	if !ok {
	// 		s2, ok := p2.(string)

	// 		if ok {
	// 			c2, errT = tk.StrToIntQuick(s2)

	// 			if errT != nil {
	// 				return p.ErrStrf("failed to convert to int: %v", errT)
	// 			}
	// 		} else {
	// 			c2 = tk.ToInt(p2)
	// 		}
	// 	}

	// 	p.Push(c1 < c2)

	// 	return ""

	// case 721: // <i$
	// 	p.Push(p.Pop().(int) > p.Pop().(int))

	// 	return ""

	// case 722: // <i*
	// 	regsT := p.CurrentFuncContextM.RegsM
	// 	regsT.CondsM[0] = regsT.IntsM[0] < regsT.IntsM[1]

	// 	return ""

	case 790: // cmp/比较
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		var v3 int

		switch nv := v2.(type) {
		case bool:
			v1v := v1.(bool)

			if v1v == nv {
				v3 = 0
			} else if v1v == false {
				v3 = -1
			} else {
				v3 = 1
			}
		case int:
			v1v := v1.(int)

			if v1v == nv {
				v3 = 0
			} else if v1v < nv {
				v3 = -1
			} else {
				v3 = 1
			}
		case byte:
			v1v := v1.(byte)

			if v1v == nv {
				v3 = 0
			} else if v1v < nv {
				v3 = -1
			} else {
				v3 = 1
			}
		case rune:
			v1v := v1.(rune)

			if v1v == nv {
				v3 = 0
			} else if v1v < nv {
				v3 = -1
			} else {
				v3 = 1
			}
		case float64:
			v1v := v1.(float64)

			if v1v == nv {
				v3 = 0
			} else if v1v < nv {
				v3 = -1
			} else {
				v3 = 1
			}
		case string:
			v1v := v1.(string)

			if v1v == nv {
				v3 = 0
			} else if v1v < nv {
				v3 = -1
			} else {
				v3 = 1
			}
		default:
			return p.ErrStrf("数据类型不匹配")
		}

		p.SetVarInt(pr, v3)

		return ""
	case 801: // inc
		if instrT.ParamLen < 1 {
			v1 := p.Pop()

			nv, ok := v1.(int)

			if ok {
				p.Push(nv + 1)
				return ""
			}

			nv2, ok := v1.(byte)

			if ok {
				p.Push(nv2 + 1)
				return ""
			}

			nv3, ok := v1.(rune)

			if ok {
				p.Push(nv3 + 1)
				return ""
			}

			p.Push(tk.ToInt(v1) + 1)

			return ""
		}

		// varsT := (*(p.CurrentVarsM))

		v1 := p.GetVarValue(instrT.Params[0])
		// v1 := varsT[p1].(int)

		nv, ok := v1.(int)

		if ok {
			p.SetVarInt(instrT.Params[0].Ref, nv+1)
			return ""
		}

		nv2, ok := v1.(byte)

		if ok {
			p.SetVarInt(instrT.Params[0].Ref, nv2+1)
			return ""
		}

		nv3, ok := v1.(rune)

		if ok {
			p.SetVarInt(instrT.Params[0].Ref, nv3+1)
			return ""
		}

		p.SetVarInt(instrT.Params[0].Ref, tk.ToInt(v1)+1)

		// varsT[p1] = v1 + 1

		return ""

	// case 803: // inc*
	// 	v1 := instrT.Params[0].Value.(int)

	// 	p.CurrentFuncContextM.RegsM.IntsM[v1]++

	// 	return ""

	case 810: // dec
		if instrT.ParamLen < 1 {
			v1 := p.Pop()

			nv, ok := v1.(int)

			if ok {
				p.Push(nv - 1)
				return ""
			}

			nv2, ok := v1.(byte)

			if ok {
				p.Push(nv2 - 1)
				return ""
			}

			nv3, ok := v1.(rune)

			if ok {
				p.Push(nv3 - 1)
				return ""
			}

			p.Push(tk.ToInt(v1) - 1)

			return ""
		}

		// varsT := (*(p.CurrentVarsM))

		v1 := p.GetVarValue(instrT.Params[0])
		// v1 := varsT[p1].(int)

		nv, ok := v1.(int)

		if ok {
			p.SetVarInt(instrT.Params[0].Ref, nv-1)
			return ""
		}

		nv2, ok := v1.(byte)

		if ok {
			p.SetVarInt(instrT.Params[0].Ref, nv2-1)
			return ""
		}

		nv3, ok := v1.(rune)

		if ok {
			p.SetVarInt(instrT.Params[0].Ref, nv3-1)
			return ""
		}

		p.SetVarInt(instrT.Params[0].Ref, tk.ToInt(v1)-1)

		// varsT[p1] = v1 + 1

		return ""

	// case 811: // dec$
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	// varsT := (*(p.CurrentVarsM))

	// 	p1 := instrT.Params[0].Ref
	// 	// v1 := p.GetVarValue(instrT.Params[0])
	// 	v1 := p.GetVarValue(instrT.Params[0]).(int)

	// 	// if tk.IsError(v1) {
	// 	// 	return p.ErrStrf("invalid param: %v", v1)
	// 	// }

	// 	p.SetVarInt(p1, v1-1)
	// 	// varsT[p1] = v1 - 1

	// 	return ""

	// case 812: // dec*
	// 	v1 := instrT.Params[0].Value.(int)

	// 	p.CurrentFuncContextM.RegsM.IntsM[v1]--

	// 	return ""

	// case 820: // intAdd
	// 	pr := -5
	// 	v1p := 0

	// 	// if instrT.ParamLen > 2 {
	// 	// 	pr = instrT.Params[0].Ref
	// 	// 	v1p = 1
	// 	// }

	// 	var v1, v2 interface{}

	// 	if instrT.ParamLen == 0 {
	// 		v2 = p.Pop()
	// 		v1 = p.Pop()
	// 	} else if instrT.ParamLen == 1 {
	// 		pr = instrT.Params[0].Ref

	// 		v2 = p.Pop()
	// 		v1 = p.Pop()
	// 	} else if instrT.ParamLen == 2 {
	// 		v1 = p.GetVarValue(instrT.Params[v1p])
	// 		v2 = p.GetVarValue(instrT.Params[v1p+1])
	// 	} else {
	// 		pr = instrT.Params[0].Ref
	// 		v1p = 1

	// 		v1 = p.GetVarValue(instrT.Params[v1p])
	// 		v2 = p.GetVarValue(instrT.Params[v1p+1])
	// 	}

	// 	p.SetVarInt(pr, tk.ToInt(v1)+tk.ToInt(v2))

	// 	return ""

	// case 821: // intAdd$
	// 	p.Push(p.Pop().(int) + p.Pop().(int))

	// 	return ""

	// case 831: // intDiv
	// 	if instrT.ParamLen < 2 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	v1 := p.GetVarValue(instrT.Params[0])

	// 	v2 := p.GetVarValue(instrT.Params[1])

	// 	p.Push(tk.ToInt(v1) / tk.ToInt(v2))

	// 	return ""

	// case 840: // floatAdd
	// 	if instrT.ParamLen < 2 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	v1 := p.GetVarValue(instrT.Params[0])

	// 	v2 := p.GetVarValue(instrT.Params[1])

	// 	p.Push(tk.ToFloat(v1) + tk.ToFloat(v2))

	// 	return ""

	// case 848: // floatDiv
	// 	if instrT.ParamLen < 2 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	v1 := p.GetVarValue(instrT.Params[0])

	// 	v2 := p.GetVarValue(instrT.Params[1])

	// 	p.Push(tk.ToFloat(v1) / tk.ToFloat(v2))

	// 	return ""

	case 901: // add
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		var v3 interface{}

		switch nv := v1.(type) {
		case int:
			v3 = nv + v2.(int)
		case byte:
			v3 = nv + v2.(byte)
		case rune:
			v3 = nv + v2.(rune)
		case float64:
			v3 = nv + v2.(float64)
		case string:
			v3 = nv + v2.(string)
		case time.Time:
			v3 = nv.Add(time.Duration(time.Millisecond * time.Duration(tk.ToInt(v2))))
		default:
			return p.ErrStrf("数据类型不匹配：%T(%v) -> %T(%v)", v1, v1, v2, v2)
		}

		p.SetVarInt(pr, v3)

		return ""
	case 902: // sub/-/减
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		var v3 interface{}

		switch nv := v1.(type) {
		case int:
			v3 = nv - v2.(int)
		case byte:
			v3 = nv - v2.(byte)
		case rune:
			v3 = nv - v2.(rune)
		case float64:
			v3 = nv - v2.(float64)
		case time.Time:
			rsT := tk.ToTime(v2)

			if tk.IsError(rsT) {
				t2 := tk.ToInt(v2, tk.MAX_INT)

				if t2 == tk.MAX_INT {
					return p.ErrStrf("时间转换失败：%T -> %T", v1, v2)
				}

				v3 = nv.Add(time.Duration(-t2) * time.Millisecond)
			} else {
				v3 = tk.ToInt(nv.Sub(rsT.(time.Time)) / time.Millisecond)
			}

		default:
			return p.ErrStrf("数据类型不匹配")
		}

		p.SetVarInt(pr, v3)

		return ""
	case 903: // mul/*/乘
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		var v3 interface{}

		switch nv := v1.(type) {
		case int:
			v3 = nv * v2.(int)
		case byte:
			v3 = nv * v2.(byte)
		case rune:
			v3 = nv * v2.(rune)
		case float64:
			v3 = nv * v2.(float64)
		default:
			return p.ErrStrf("数据类型不匹配")
		}

		p.SetVarInt(pr, v3)

		return ""
	case 904: // div///除
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		var v3 interface{}

		switch nv := v1.(type) {
		case int:
			v3 = nv / v2.(int)
		case byte:
			v3 = nv / v2.(byte)
		case rune:
			v3 = nv / v2.(rune)
		case float64:
			v3 = nv / v2.(float64)
		case time.Duration:
			v3 = int(nv) / tk.ToInt(v2)
		default:
			return p.ErrStrf("数据类型不匹配：%T -> %T", nv, v2)
		}

		p.SetVarInt(pr, v3)

		return ""
	case 905: // mod/%/取模
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		var v3 interface{}

		switch nv := v1.(type) {
		case int:
			v3 = nv % v2.(int)
		case byte:
			v3 = nv % v2.(byte)
		case rune:
			v3 = nv % v2.(rune)
		default:
			return p.ErrStrf("数据类型不匹配")
		}

		p.SetVarInt(pr, v3)

		return ""
	case 930: // !
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5

		var v2 interface{}

		if instrT.ParamLen < 1 {
			v2 = p.Pop()
		} else if instrT.ParamLen < 2 {
			v2 = p.GetVarValue(instrT.Params[0])
		} else {
			pr = instrT.Params[0].Ref
			v2 = p.GetVarValue(instrT.Params[1])
		}

		var v3 interface{}

		switch nv := v2.(type) {
		case bool:
			v3 = !nv
		default:
			if v2 == nil {
				v3 = true
			} else if v2 == Undefined {
				v3 = true
			} else {
				v3 = false
			}
		}

		p.SetVarInt(pr, v3)

		return ""
	case 931: // not
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5

		var v2 interface{}

		if instrT.ParamLen < 1 {
			v2 = p.Pop()
		} else if instrT.ParamLen < 2 {
			v2 = p.GetVarValue(instrT.Params[0])
		} else {
			pr = instrT.Params[0].Ref
			v2 = p.GetVarValue(instrT.Params[1])
		}

		var v3 interface{}

		switch nv := v2.(type) {
		case bool:
			v3 = !nv
		case byte:
			v3 = ^nv
		case rune:
			v3 = ^nv
		case int:
			v3 = ^nv
		case string:
			buf, err := hex.DecodeString(nv)
			if err != nil {
				return p.ErrStrf("16进制转换错误")
			}

			for i := 0; i < len(buf); i++ {
				buf[i] = ^(buf[i])
			}

			v3 = hex.EncodeToString(buf)
		default:
			return p.ErrStrf("参数类型不匹配")
		}

		p.SetVarInt(pr, v3)

		return ""
	case 933: // &&
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v2 := p.GetVarValue(instrT.Params[v1p]).(bool)

		v3 := p.GetVarValue(instrT.Params[v1p+1]).(bool)

		p.SetVarInt(pr, v2 && v3)

		return ""
	case 934: // ||
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v2 := p.GetVarValue(instrT.Params[v1p]).(bool)

		v3 := p.GetVarValue(instrT.Params[v1p+1]).(bool)

		p.SetVarInt(pr, v2 || v3)

		return ""
	case 990: // ? 三元操作符
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数个数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToBool(p.GetVarValue(instrT.Params[v1p]))

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		v3 := p.GetVarValue(instrT.Params[v1p+2])

		if v1 {
			p.SetVarInt(pr, v2)
		} else {
			p.SetVarInt(pr, v3)
		}

		return ""
	case 998: // eval
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数个数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, p.EvalExpression(v1))

		return ""
	case 999: // quickEval
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数个数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		// tk.Plv(instrT)

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p])) //instrT.Line

		p.SetVarInt(pr, p.QuickEval(v1))

		return ""
	case 1010: // call
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		tmpPointerT, ok := v1.(int)

		if !ok {
			tmps, ok := v1.(string)

			if !ok {
				return p.ErrStrf("参数类型错误")
			}

			if !strings.HasPrefix(tmps, ":") {
				return p.ErrStrf("标号格式错误：%v", tmps)
			}

			tmps = tmps[1:]

			varIndexT, ok := p.VarIndexMapM[tmps]

			if !ok {
				return p.ErrStrf("无效的标号：%v", tmps)
			}

			tmpPointerT, ok = p.LabelsM[varIndexT]

			if !ok {
				return p.ErrStrf("无效的标号序号：%v(%v)", varIndexT, tmps)
			}

			p.InstrListM[lineA].Params[0].Value = tmpPointerT
			// instrT = VarRef{Ref: instrT.Params[0].Ref, Value: tmpPointerT}
			// tk.Plv(instrT.Params[0])
		}
		// p1 := instrT.Params[0].Value.(int)

		// tk.Pln(tk.ToJSONX(p, "-indent", "-sort"))
		// tk.Pln("p1", p1)
		// tk.Exit()
		p.PushFunc()

		return tmpPointerT

	case 1020: // ret
		rsi := p.RunDefer()

		if tk.IsErrX(rsi) {
			return p.ErrStrf("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStrX(rsi))
		}

		pT := p.PopFunc()

		return pT

	case 1050: // callFunc
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		argCountT := 0

		codeT := ""

		if instrT.ParamLen > 1 {
			argCountT = tk.ToInt(p.GetVarValue(instrT.Params[0]))

			codeT = p.GetVarValue(instrT.Params[1]).(string)
		} else {
			codeT = p.GetVarValue(instrT.Params[0]).(string)
		}

		codeT = strings.ReplaceAll(codeT, "~~~", "`")

		return p.CallFunc(codeT, argCountT)

	case 1060: // goFunc
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		argCountT := 0

		codeT := ""

		if instrT.ParamLen > 1 {
			argCountT = tk.ToInt(p.GetVarValue(instrT.Params[0]))

			codeT = p.GetVarValue(instrT.Params[1]).(string)
		} else {
			codeT = p.GetVarValue(instrT.Params[0]).(string)
		}

		codeT = strings.ReplaceAll(codeT, "~~~", "`")

		return p.GoFunc(codeT, argCountT)

	case 1070: // fastCall
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pointerT := p.CodePointerM

		v1 := p.GetVarValue(instrT.Params[0])

		for ii := 0; ii < (instrT.ParamLen - 1); ii++ {
			p.Push(p.GetVarValue(instrT.Params[ii+1]))
		}

		tmpPointerT, ok := v1.(int)

		if !ok {
			tmps, ok := v1.(string)

			if !ok {
				return p.ErrStrf("参数类型错误")
			}

			if !strings.HasPrefix(tmps, ":") {
				return p.ErrStrf("标号格式错误：%v", tmps)
			}

			tmps = tmps[1:]

			varIndexT, ok := p.VarIndexMapM[tmps]

			if !ok {
				return p.ErrStrf("无效的标号：%v", tmps)
			}

			tmpPointerT, ok = p.LabelsM[varIndexT]

			if !ok {
				return p.ErrStrf("无效的标号序号：%v(%v)", varIndexT, tmps)
			}

			p.InstrListM[lineA].Params[0].Value = tmpPointerT
			// instrT = VarRef{Ref: instrT.Params[0].Ref, Value: tmpPointerT}
			// tk.Plv(instrT.Params[0])
		}

		for {
			rs := p.RunLine(tmpPointerT)

			nv, ok := rs.(int)

			if ok {
				tmpPointerT = nv
				continue
			}

			nsv, ok := rs.(string)

			if ok {
				if tk.IsErrStr(nsv) {
					return nsv
				}

				if nsv == "exit" {
					return "exit"
				} else if nsv == "fr" {
					break
				}
			}

			tmpPointerT++
		}

		return pointerT + 1

	case 1071: // fastRet
		return "fr"

	case 1080: // for
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		p1 := instrT.Params[0]

		// v1 := p.GetVarValue(instrT.Params[0])
		labelT := tk.ToInt(p.GetVarValue(instrT.Params[1]))

		pointerT := p.CodePointerM

		startPointerT := labelT

	for51:
		for true {
			v1 := tk.ToBool(p.GetVarValue(p1))
			// tk.Pl("v1: %v, %v, %#v", v1, p.GetVarValue(p1), p1)
			if !v1 {
				break for51
			}

			tmpPointerT := startPointerT

			if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
				return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
			}

			for {
				// tk.Pln(i, v)
				// tk.Pl("%v %v %v", lineA, len(p.InstrListM), tk.LimitString(p.SourceM[p.CodeSourceMapM[p.CodePointerM]], 50))
				// tk.Pl("%v %v %v", tmpPointerT, len(p.InstrListM), tk.LimitString(p.SourceM[p.CodeSourceMapM[tmpPointerT]], 50))

				// tk.Exit()

				rs := p.RunLine(tmpPointerT)

				nv, ok := rs.(int)

				if ok {
					tmpPointerT = nv
					continue
				}

				nsv, ok := rs.(string)

				if ok {
					if tk.IsErrStr(nsv) {
						return nsv
					}

					if nsv == "exit" {
						return "exit"
					} else if nsv == "cont" {

						continue for51
					} else if nsv == "brk" {
						break for51
					}
				}

				tmpPointerT++
			}

		}

		// p.CodePointerM = pointerT

		return pointerT + 1

	case 1085: // range
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		labelPT := 1

		v1 := p.GetVarValue(instrT.Params[0])
		var v2 interface{} = nil

		if instrT.ParamLen > 2 {
			labelPT = 2
			v2 = p.GetVarValue(instrT.Params[1])
		}

		labelT := tk.ToInt(p.GetVarValue(instrT.Params[labelPT]))

		pointerT := p.CodePointerM

		startPointerT := labelT

		// tmpPointerT := labelT

		var lenT int

		switch nv := v1.(type) {
		case []interface{}:
			lenT = len(nv)
		case []bool:
			lenT = len(nv)
		case int:
			if v2 != nil {
				lenT = tk.ToInt(v2) - nv
			} else {
				lenT = nv
			}
		case string:
			lenT = len(nv)
		case []rune:
			lenT = len(nv)
		case []byte:
			lenT = len(nv)
		case []int64:
			lenT = len(nv)
		case []float64:
			lenT = len(nv)
		case []string:
			lenT = len(nv)
		case []map[string]string:
			lenT = len(nv)
		case []map[string]interface{}:
			lenT = len(nv)

			// maps start

		case map[string]interface{}:
		for31a:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for31a
						} else if nsv == "brk" {
							break for31a
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		case map[string]int:
		for32a:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for32a
						} else if nsv == "brk" {
							break for32a
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		case map[string]byte:
		for35a:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for35a
						} else if nsv == "brk" {
							break for35a
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		case map[string]rune:
		for36a:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for36a
						} else if nsv == "brk" {
							break for36a
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		case map[string]float64:
		for33a:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for33a
						} else if nsv == "brk" {
							break for33a
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		case map[string]string:
		for34a:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for34a
						} else if nsv == "brk" {
							break for34a
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1

			// maps end
		default:
			return p.ErrStrf("参数类型错误：%T(%v)", v1, nv)
		}

	for21:
		for i := 0; i < lenT; i++ {
			// for i, v := range v1 {
			tmpPointerT := startPointerT

			if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
				return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
			}

			p.CodePointerM = tmpPointerT

			switch nv := v1.(type) {
			case []interface{}:
				p.Push(nv[i])
			case []bool:
				p.Push(nv[i])
			case int:
				if v2 != nil {
					p.Push(nv + i)
				} else {
					p.Push(i)
				}
			case string:
				p.Push(string(nv[i]))
			case []byte:
				p.Push(nv[i])
			case []rune:
				p.Push(nv[i])
			case []int64:
				p.Push(nv[i])
			case []float64:
				p.Push(nv[i])
			case []string:
				p.Push(nv[i])
			case []map[string]string:
				p.Push(nv[i])
			case []map[string]interface{}:
				p.Push(nv[i])
			default:
				return p.ErrStrf("参数类型错误：%T(%v)", v1, nv)
			}
			// p.Push(v)
			p.Push(i)

			for {
				// tk.Pln(i, v)
				// tk.Pl("%v %v %v", lineA, len(p.InstrListM), tk.LimitString(p.SourceM[p.CodeSourceMapM[p.CodePointerM]], 50))
				// tk.Pl("%v %v %v", tmpPointerT, len(p.InstrListM), tk.LimitString(p.SourceM[p.CodeSourceMapM[tmpPointerT]], 50))

				// tk.Exit()

				rs := p.RunLine(tmpPointerT)

				nv, ok := rs.(int)

				if ok {
					tmpPointerT = nv
					p.CodePointerM = tmpPointerT
					continue
				}

				nsv, ok := rs.(string)

				if ok {
					if tk.IsErrStr(nsv) {
						return nsv
					}

					if nsv == "exit" {
						return "exit"
					} else if nsv == "cont" {

						continue for21
					} else if nsv == "brk" {
						break for21
					}
				}

				tmpPointerT++
				p.CodePointerM = tmpPointerT
			}

		}

		// p.CodePointerM = pointerT

		return pointerT + 1

	case 1101: // newList/newArray
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		vs := p.ParamsToList(instrT, v1p)

		vr := make([]interface{}, 0, len(vs))

		for _, v := range vr {
			vr = append(vr, v)
		}

		p.SetVarInt(pr, vr)

		return ""

	case 1110: // addItem/addArrayItem/addListItem
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		// varsT := (*(p.CurrentVarsM))

		// p1 := instrT.Params[0].Ref
		p1 := p.GetVarRef(instrT.Params[0])

		v1 := *p1

		// tk.Plv(p1)

		// tk.Pln(p1, instrT, p)
		// tk.Pl("[Line: %v] STACK: %v, CONTEXT: %v", p.CodeSourceMapM[p.CodePointerM]+1, p.StackM[:p.StackPointerM], tk.ToJSONX(p.CurrentFuncContextM, "-inddent", "-sort"))

		// v2 := p.GetVarValue(instrT.Params[1])

		// varsT := p.GetVarValue(instrT.Params[1])

		// tk.Pln(p1, v2, varsT[p1])
		// varsT[p1] = append((varsT[p1]).([]interface{}), v2)

		var v2 interface{}

		if instrT.ParamLen < 2 {
			v2 = p.Pop()
		} else {
			v2 = p.GetVarValue(instrT.Params[1])
		}
		// *p1 = append((*p1).([]interface{}), v2)

		switch nv := v1.(type) {
		case []interface{}:
			*p1 = append((*p1).([]interface{}), v2)
		case []bool:
			*p1 = append(nv, tk.ToBool(v2))
		case []int:
			*p1 = append(nv, tk.ToInt(v2))
		case []byte:
			*p1 = append(nv, byte(tk.ToInt(v2)))
		case []rune:
			*p1 = append(nv, rune(tk.ToInt(v2)))
		case []int64:
			*p1 = append(nv, int64(tk.ToInt(v2)))
		case []float64:
			*p1 = append(nv, tk.ToFloat(v2))
		case []string:
			*p1 = append(nv, tk.ToStr(v2))
		default:
			return p.ErrStrf("参数类型错误：%T(%v) -> %T", v1, nv, v2)
			// tk.Pln(p.ErrStrf("参数类型：%T(%v) -> %T", nv, nv, v2))
		}

		return ""

	case 1111: // addStrItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		// varsT := p.GetVars()

		// p1 := instrT.Params[0].Ref
		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		// varsT[p1] = append((varsT[p1]).([]string), tk.ToStr(v2))
		v1 = append(v1.([]string), tk.ToStr(v2))

		return ""

	case 1112: // deleteItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		// varsT := (*(p.CurrentVarsM))

		// p1 := instrT.Params[0].Ref

		// p1 := p.GetVarValue(instrT.Params[0]) // instrT.Params[0].Ref
		p1 := p.GetVarRef(instrT.Params[0])

		v2 := tk.ToInt(p.GetVarValue(instrT.Params[1]))

		// varsT := p.GetVars()

		// aryT := (varsT[p1]).([]interface{})
		v1 := *p1

		switch nv := v1.(type) {
		case []interface{}:
			aryT := (*p1).([]interface{})

			if v2 >= len(aryT) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(aryT))
			}

			rs := make([]interface{}, 0, len(aryT)-1)
			rs = append(rs, aryT[:v2]...)
			rs = append(rs, aryT[v2+1:]...)

			(*p1) = rs

		case []bool:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			rs := make([]bool, 0, len(nv)-1)
			rs = append(rs, nv[:v2]...)
			rs = append(rs, nv[v2+1:]...)

			(*p1) = rs

		case []int:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			rs := make([]int, 0, len(nv)-1)
			rs = append(rs, nv[:v2]...)
			rs = append(rs, nv[v2+1:]...)

			(*p1) = rs
		case []byte:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			rs := make([]byte, 0, len(nv)-1)
			rs = append(rs, nv[:v2]...)
			rs = append(rs, nv[v2+1:]...)

			(*p1) = rs
		case []rune:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			rs := make([]rune, 0, len(nv)-1)
			rs = append(rs, nv[:v2]...)
			rs = append(rs, nv[v2+1:]...)

			(*p1) = rs
		case []int64:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			rs := make([]int64, 0, len(nv)-1)
			rs = append(rs, nv[:v2]...)
			rs = append(rs, nv[v2+1:]...)

			(*p1) = rs
		case []float64:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			rs := make([]float64, 0, len(nv)-1)
			rs = append(rs, nv[:v2]...)
			rs = append(rs, nv[v2+1:]...)

			(*p1) = rs
		case []string:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			rs := make([]string, 0, len(nv)-1)
			rs = append(rs, nv[:v2]...)
			rs = append(rs, nv[v2+1:]...)

			(*p1) = rs
		default:
			return p.ErrStrf("参数类型错误：%T(%v) -> %T", v1, nv, v2)
		}

		return ""

	case 1115: // addItems
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		// varsT := (*(p.CurrentVarsM))

		pr := instrT.Params[0].Ref
		v1p := 0

		if instrT.ParamLen > 2 {
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		switch nv := v1.(type) {
		case []interface{}:
			p.SetVarInt(pr, append(nv, (v2.([]interface{}))...))
		case []bool:
			p.SetVarInt(pr, append(nv, (v2.([]bool))...))
		case []int:
			p.SetVarInt(pr, append(nv, (v2.([]int))...))
		case []byte:
			p.SetVarInt(pr, append(nv, (v2.([]byte))...))
		case []rune:
			p.SetVarInt(pr, append(nv, (v2.([]rune))...))
		case []int64:
			p.SetVarInt(pr, append(nv, (v2.([]int64))...))
		case []float64:
			p.SetVarInt(pr, append(nv, (v2.([]float64))...))
		case []string:
			p.SetVarInt(pr, append(nv, (v2.([]string))...))
		default:
			return p.ErrStrf("参数类型错误：%T(%v) -> %T", v1, nv, v2)
			// tk.Pln(p.ErrStrf("参数类型：%T(%v) -> %T", nv, nv, v2))
		}

		// if instrT.ParamLen > 2 {
		// 	// p.SetVarInt(instrT.Params[2].Ref, append((varsT[p1]).([]interface{}), v2...))
		// 	p.SetVarInt(instrT.Params[2].Ref, append(v1.([]interface{}), v2...))
		// } else {
		// 	// varsT[p1] = append((varsT[p1]).([]interface{}), v2...)
		// 	v1 = append(v1.([]interface{}), v2...)
		// }

		return ""

	case 1123: // getItem/getArrayItem
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		var pr = instrT.Params[0].Ref

		v1 := p.GetVarValue(instrT.Params[1])

		v2 := tk.ToInt(p.GetVarValue(instrT.Params[2]))

		switch nv := v1.(type) {
		case []interface{}:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVarInt(pr, p.GetVarValue(instrT.Params[3]))
					return ""
				} else {
					return p.ErrStrf("索引超出范围：%v/%v", v2, len(nv))
				}
			}
			p.SetVarInt(pr, nv[v2])
		case []bool:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVarInt(pr, p.GetVarValue(instrT.Params[3]))
					return ""
				} else {
					return p.ErrStrf("索引超出范围：%v/%v", v2, len(nv))
				}
			}
			p.SetVarInt(pr, nv[v2])
		case []int:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVarInt(pr, p.GetVarValue(instrT.Params[3]))
					return ""
				} else {
					return p.ErrStrf("索引超出范围：%v/%v", v2, len(nv))
				}
			}
			p.SetVarInt(pr, nv[v2])
		case []byte:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVarInt(pr, p.GetVarValue(instrT.Params[3]))
					return ""
				} else {
					return p.ErrStrf("索引超出范围：%v/%v", v2, len(nv))
				}
			}
			p.SetVarInt(pr, nv[v2])
		case []rune:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVarInt(pr, p.GetVarValue(instrT.Params[3]))
					return ""
				} else {
					return p.ErrStrf("索引超出范围：%v/%v", v2, len(nv))
				}
			}
			p.SetVarInt(pr, nv[v2])
		case []int64:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVarInt(pr, p.GetVarValue(instrT.Params[3]))
					return ""
				} else {
					return p.ErrStrf("索引超出范围：%v/%v", v2, len(nv))
				}
			}
			p.SetVarInt(pr, nv[v2])
		case []float64:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVarInt(pr, p.GetVarValue(instrT.Params[3]))
					return ""
				} else {
					return p.ErrStrf("索引超出范围：%v/%v", v2, len(nv))
				}
			}
			p.SetVarInt(pr, nv[v2])
		case []string:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVarInt(pr, p.GetVarValue(instrT.Params[3]))
					return ""
				} else {
					return p.ErrStrf("索引超出范围：%v/%v", v2, len(nv))
				}
			}
			p.SetVarInt(pr, nv[v2])
		case []map[string]string:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVarInt(pr, p.GetVarValue(instrT.Params[3]))
					return ""
				} else {
					return p.ErrStrf("索引超出范围：%v/%v", v2, len(nv))
				}
			}

			p.SetVarInt(pr, nv[v2])
		default:
			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, p.GetVarValue(instrT.Params[3]))
			} else {
				p.SetVarInt(pr, Undefined)
			}
			return p.ErrStrf("参数类型错误")
		}

		return ""

	case 1124: // setItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := tk.ToInt(p.GetVarValue(instrT.Params[1]))

		var v3 interface{}

		if instrT.ParamLen < 3 {
			v3 = p.Pop()
			// v1.([]interface{})[v2] = p.Pop()
		} else {
			v3 = p.GetVarValue(instrT.Params[2])
			// v1.([]interface{})[v2] = p.GetVarValue(instrT.Params[2])
		}

		switch nv := v1.(type) {
		case []interface{}:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = v3
		case []bool:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = tk.ToBool(v3)
		case []int:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = tk.ToInt(v3)
		case []byte:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = byte(tk.ToInt(v3))
		case []rune:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = rune(tk.ToInt(v3))
		case []int64:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = int64(tk.ToInt(v3))
		case []float64:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = tk.ToFloat(v3)
		case []string:
			if v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = tk.ToStr(v3)
		default:
			return p.ErrStrf("参数类型错误")

		}

		return ""

	case 1130: // slice
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2vT := p.GetVarValue(instrT.Params[v1p+1])
		v3vT := p.GetVarValue(instrT.Params[v1p+2])

		v2 := tk.ToInt(v2vT, -1)
		v3 := tk.ToInt(v3vT, -1)

		// varsT := p.GetVars()
		switch nv := v1.(type) {
		case []interface{}:
			if v2 == -1 {
				if tk.ToStr(v2vT) == "-" {
					v2 = 0
				}
			}

			if v3 < 0 {
				if tk.ToStr(v3vT) == "-" {
					v3 = len(nv)
				}
			}

			if v2 < 0 || v2 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []bool:
			if v2 == -1 {
				if tk.ToStr(v2vT) == "-" {
					v2 = 0
				}
			}

			if v3 < 0 {
				if tk.ToStr(v3vT) == "-" {
					v3 = len(nv)
				}
			}

			if v2 < 0 || v2 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []int:
			if v2 == -1 {
				if tk.ToStr(v2vT) == "-" {
					v2 = 0
				}
			}

			if v3 < 0 {
				if tk.ToStr(v3vT) == "-" {
					v3 = len(nv)
				}
			}

			if v2 < 0 || v2 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []byte:
			if v2 == -1 {
				if tk.ToStr(v2vT) == "-" {
					v2 = 0
				}
			}

			if v3 < 0 {
				if tk.ToStr(v3vT) == "-" {
					v3 = len(nv)
				}
			}

			if v2 < 0 || v2 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []rune:
			if v2 == -1 {
				if tk.ToStr(v2vT) == "-" {
					v2 = 0
				}
			}

			if v3 < 0 {
				if tk.ToStr(v3vT) == "-" {
					v3 = len(nv)
				}
			}

			if v2 < 0 || v2 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []int64:
			if v2 == -1 {
				if tk.ToStr(v2vT) == "-" {
					v2 = 0
				}
			}

			if v3 < 0 {
				if tk.ToStr(v3vT) == "-" {
					v3 = len(nv)
				}
			}

			if v2 < 0 || v2 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []float64:
			if v2 == -1 {
				if tk.ToStr(v2vT) == "-" {
					v2 = 0
				}
			}

			if v3 < 0 {
				if tk.ToStr(v3vT) == "-" {
					v3 = len(nv)
				}
			}

			if v2 < 0 || v2 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []string:
			if v2 == -1 {
				if tk.ToStr(v2vT) == "-" {
					v2 = 0
				}
			}

			if v3 < 0 {
				if tk.ToStr(v3vT) == "-" {
					v3 = len(nv)
				}
			}

			if v2 < 0 || v2 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case string:
			if v2 == -1 {
				if tk.ToStr(v2vT) == "-" {
					v2 = 0
				}
			}

			if v3 < 0 {
				if tk.ToStr(v3vT) == "-" {
					v3 = len(nv)
				}
			}

			if v2 < 0 || v2 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		default:
			return p.ErrStrf("参数类型错误：%T(%v) -> %T", v1, nv, v2)
			// tk.Pln(p.ErrStrf("参数类型：%T(%v) -> %T", nv, nv, v2))
		}

		// if instrT.ParamLen > 3 {
		// 	// p.SetVarInt(instrT.Params[3].Ref, ((varsT[p1]).([]interface{}))[v2:v3])
		// 	p.SetVarInt(instrT.Params[3].Ref, (v1.([]interface{}))[v2:v3])
		// } else {
		// 	// varsT[p1] = ((varsT[p1]).([]interface{}))[v2:v3]
		// 	v1 = (v1.([]interface{}))[v2:v3]
		// }

		return ""

	case 1140: // rangeList
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0])
		labelT := tk.ToInt(p.GetVarValue(instrT.Params[1]))

		pointerT := p.CodePointerM

		startPointerT := labelT

		// tmpPointerT := labelT

		var lenT int

		switch nv := v1.(type) {
		case []interface{}:
			lenT = len(nv)
		case []bool:
			lenT = len(nv)
		case []int:
			lenT = len(nv)
		case []byte:
			lenT = len(nv)
		case []rune:
			lenT = len(nv)
		case []int64:
			lenT = len(nv)
		case []float64:
			lenT = len(nv)
		case []string:
			lenT = len(nv)
		default:
			return p.ErrStrf("参数类型错误：%T(%v)", v1, nv)
		}

	for1:
		for i := 0; i < lenT; i++ {
			// for i, v := range v1 {
			tmpPointerT := startPointerT
			p.CodePointerM = tmpPointerT

			if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
				return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
			}

			switch nv := v1.(type) {
			case []interface{}:
				p.Push(nv[i])
			case []bool:
				p.Push(nv[i])
			case []int:
				p.Push(nv[i])
			case []byte:
				p.Push(nv[i])
			case []rune:
				p.Push(nv[i])
			case []int64:
				p.Push(nv[i])
			case []float64:
				p.Push(nv[i])
			case []string:
				p.Push(nv[i])
			default:
				return p.ErrStrf("参数类型错误：%T(%v)", v1, nv)
			}
			// p.Push(v)
			p.Push(i)

			for {
				// tk.Pln(i, v)
				// tk.Pl("%v %v %v", lineA, len(p.InstrListM), tk.LimitString(p.SourceM[p.CodeSourceMapM[p.CodePointerM]], 50))
				// tk.Pl("%v %v %v", tmpPointerT, len(p.InstrListM), tk.LimitString(p.SourceM[p.CodeSourceMapM[tmpPointerT]], 50))

				// tk.Exit()

				rs := p.RunLine(tmpPointerT)

				nv, ok := rs.(int)

				if ok {
					tmpPointerT = nv
					p.CodePointerM = tmpPointerT
					continue
				}

				nsv, ok := rs.(string)

				if ok {
					if tk.IsErrStr(nsv) {
						return nsv
					}

					if nsv == "exit" {
						return "exit"
					} else if nsv == "cont" {

						continue for1
					} else if nsv == "brk" {
						break for1
					}
				}

				tmpPointerT++
				p.CodePointerM = tmpPointerT
			}

		}

		// p.CodePointerM = pointerT

		return pointerT + 1

	case 1141: // rangeStrList
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0]).([]string)
		labelT := tk.ToInt(p.GetVarValue(instrT.Params[1]))

		pointerT := p.CodePointerM

		startPointerT := labelT

		// tmpPointerT := labelT

	for5:
		for i, v := range v1 {
			tmpPointerT := startPointerT

			if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
				return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
			}

			p.Push(v)
			p.Push(i)

			for {
				// tk.Pln(i, v)
				// tk.Pl("%v %v %v", lineA, len(p.InstrListM), tk.LimitString(p.SourceM[p.CodeSourceMapM[p.CodePointerM]], 50))
				// tk.Pl("%v %v %v", tmpPointerT, len(p.InstrListM), tk.LimitString(p.SourceM[p.CodeSourceMapM[tmpPointerT]], 50))

				// tk.Exit()

				rs := p.RunLine(tmpPointerT)

				nv, ok := rs.(int)

				if ok {
					tmpPointerT = nv
					continue
				}

				nsv, ok := rs.(string)

				if ok {
					if tk.IsErrStr(nsv) {
						return nsv
					}

					if nsv == "exit" {
						return "exit"
					} else if nsv == "cont" {

						continue for5
					} else if nsv == "brk" {
						break for5
					}
				}

				tmpPointerT++
			}

		}

		// p.CodePointerM = pointerT

		return pointerT + 1

	case 1210: // continue
		return "cont"
	case 1211: // break
		return "brk"

	case 1212: // continueIf

		var condT bool

		if instrT.ParamLen < 1 {
			condT = tk.ToBool(p.Pop())
		} else {
			condT = tk.ToBool(p.GetVarValue(instrT.Params[0]))
		}

		if condT {
			return "cont"
		}

		return ""

	case 1213: // breakIf
		// if instrT.ParamLen < 1 {
		// 	return p.ErrStrf("参数不够")
		// }

		var condT bool

		if instrT.ParamLen < 1 {
			condT = tk.ToBool(p.Pop())
		} else {
			condT = tk.ToBool(p.GetVarValue(instrT.Params[0]))
		}

		if condT {
			return "brk"
		}

		return ""

	case 1310: // setMapItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		// p1 := instrT.Params[0].Ref

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1]).(string)

		var v3 interface{}

		if instrT.ParamLen < 3 {
			v3 = p.Pop()
			// v1.(map[string]interface{})[v2] = p.Pop()
		} else {
			v3 = p.GetVarValue(instrT.Params[2])
			// v1.(map[string]interface{})[v2] = p.GetVarValue(instrT.Params[2])
		}

		switch nv := v1.(type) {
		case map[string]interface{}:
			nv[v2] = v3
		case map[string]int:
			nv[v2] = tk.ToInt(v3)
		case map[string]byte:
			nv[v2] = tk.ToByte(v3)
		case map[string]rune:
			nv[v2] = tk.ToRune(v3)
		case map[string]float64:
			nv[v2] = tk.ToFloat(v3)
		case map[string]string:
			nv[v2] = tk.ToStr(v3)
		default:
			return p.ErrStrf("参数类型错误")
		}

		return ""

	case 1312: // deleteMapItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		// varsT := (*(p.CurrentVarsM))

		// p1 := instrT.Params[0].Ref

		// p1 := p.GetVarValue(instrT.Params[0]) // instrT.Params[0].Ref
		p1 := p.GetVarRef(instrT.Params[0])

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

		// varsT := p.GetVars()

		// aryT := (varsT[p1]).([]interface{})
		// mapT := (*p1).(map[string]interface{})

		v1 := *p1

		switch nv := v1.(type) {
		case map[string]interface{}:
			delete(nv, v2)
		case map[string]int:
			delete(nv, v2)
		case map[string]byte:
			delete(nv, v2)
		case map[string]rune:
			delete(nv, v2)
		case map[string]float64:
			delete(nv, v2)
		case map[string]string:
			delete(nv, v2)
		default:
			return p.ErrStrf("参数类型错误")
		}

		// rs := make([]interface{}, 0, len(aryT)-1)
		// rs = append(rs, aryT[:v2]...)
		// rs = append(rs, aryT[v2+1:]...)

		// varsT[p1] = rs // append((varsT[p1]).([]interface{}), v2)
		// delete(mapT, v2)

		// tk.DeleteItemInArray(p1.([]interface{}), v2)

		return ""

	case 1320: // getMapItem
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref // -5
		v1p := 1

		// if instrT.ParamLen > 2 {
		// 	pr = instrT.Params[0].Ref
		// 	v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// var vT interface{}
		// tk.Pln(pr, v1, v2)

		var rv interface{}

		var ok bool

		switch nv := v1.(type) {
		case map[string]interface{}:
			rv, ok = nv[v2]
		case map[string]int:
			rv, ok = nv[v2]
		case map[string]byte:
			rv, ok = nv[v2]
		case map[string]rune:
			rv, ok = nv[v2]
		case map[string]float64:
			rv, ok = nv[v2]
		case map[string]string:
			rv, ok = nv[v2]
		default:
			return p.ErrStrf("参数类型错误")
		}

		if !ok {
			// tk.Pln("here", instrT)
			if instrT.ParamLen > 3 {
				rv = p.GetVarValue(instrT.Params[v1p+2])

			} else {
				rv = Undefined
			}
		}

		p.SetVarInt(pr, rv)

		return ""

	case 1340: // rangeMap
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0])
		labelT := tk.ToInt(p.GetVarValue(instrT.Params[1]))

		pointerT := p.CodePointerM
		// tk.Pln(p.CodeSourceMapM[pointerT]+1, tk.LimitString(p.SourceM[p.CodeSourceMapM[pointerT]], 50))

		startPointerT := labelT

		// tmpPointerT := labelT

		switch nv := v1.(type) {
		case map[string]interface{}:
		for31:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for31
						} else if nsv == "brk" {
							break for31
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		case map[string]int:
		for32:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for32
						} else if nsv == "brk" {
							break for32
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		case map[string]byte:
		for35:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for35
						} else if nsv == "brk" {
							break for35
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		case map[string]rune:
		for36:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for36
						} else if nsv == "brk" {
							break for36
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		case map[string]float64:
		for33:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for33
						} else if nsv == "brk" {
							break for33
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		case map[string]string:
		for34:
			for k, v := range nv {
				tmpPointerT := startPointerT
				p.CodePointerM = tmpPointerT

				if tmpPointerT < 0 || tmpPointerT >= len(p.CodeListM) {
					return p.ErrStrf("遍历中指令序号超出范围: %v/%v", tmpPointerT, len(p.CodeListM))
				}

				p.Push(v)
				p.Push(k)

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						p.CodePointerM = tmpPointerT
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							return nsv
						}

						if nsv == "exit" {
							return "exit"
						} else if nsv == "cont" {

							continue for34
						} else if nsv == "brk" {
							break for34
						}
					}

					tmpPointerT++
					p.CodePointerM = tmpPointerT
				}

			}

			return pointerT + 1
		default:
			return p.ErrStrf("参数类型错误")
		}

	case 1401: // new
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		// pr := -5
		// v1p := 0

		// if instrT.ParamLen > 0 {
		pr := instrT.Params[0].Ref
		v1p := 1
		// }

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		switch v1 {
		case "bool":
			p.SetVarInt(pr, new(bool))
		case "int":
			p.SetVarInt(pr, new(int))
		case "uint64":
			p.SetVarInt(pr, new(uint64))
		case "byte":
			p.SetVarInt(pr, new(byte))
		case "rune":
			p.SetVarInt(pr, new(rune))
		case "str", "string":
			p.SetVarInt(pr, new(string))
		case "byteList": // 后面可接多个字节，其中可以有字节数组（会逐一加入字节列表中）
			blT := make([]byte, 0)

			vs := p.ParamsToList(instrT, v1p+1)

			for _, vvv := range vs {
				nv, ok := vvv.([]byte)

				if ok {
					for _, vvvj := range nv {
						blT = append(blT, vvvj)
					}
				} else {
					blT = append(blT, tk.ToByte(vvv, 0))
				}
			}

			p.SetVarInt(pr, &blT)
		case "bytesBuffer", "bytesBuf":
			p.SetVarInt(pr, new(bytes.Buffer))
		case "stringBuffer", "strBuf", "strings.Builder":
			p1 := new(strings.Builder)
			if instrT.ParamLen > 2 {
				p1.WriteString(tk.ToStr(p.GetVarValue(instrT.Params[2])))
			}

			p.SetVarInt(pr, p1)
		case "time":
			timeT := time.Now()
			p.SetVarInt(pr, &timeT)
		case "mutex", "lock":
			p.SetVarInt(pr, new(sync.RWMutex))
		case "mux":
			p.SetVarInt(pr, http.NewServeMux())
		case "ssh":
			v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
			v3 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))
			v4 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+3]))
			v5 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+4]))
			if strings.HasPrefix(v5, "740404") {
				v5 = tk.DecryptStringByTXDEF(v5)
			}

			sshT, errT := tk.NewSSHClient(v2, v3, v4, v5)

			if errT != nil {
				p.SetVarInt(pr, errT)

				return ""
			}

			p.SetVarInt(pr, sshT)
		case "gui":
			objT := p.GetVar("guiG")
			p.SetVarInt(pr, objT)
		case "quickDelegate":
			if instrT.ParamLen < 2 {
				return p.ErrStrf("参数不够")
			}

			v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))

			var deleT tk.QuickDelegate

			// same as fastCall
			deleT = func(strA string) string {
				pointerT := p.CodePointerM

				p.Push(strA)

				tmpPointerT := v2

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							// tmpRs := p.Pop()
							p.CodePointerM = pointerT
							return nsv
						}

						if nsv == "exit" { // 不应发生
							tmpRs := p.Pop()
							p.CodePointerM = pointerT
							return tk.ToStr(tmpRs)
						} else if nsv == "fr" {
							break
						}
					}

					tmpPointerT++
				}

				// return pointerT + 1

				tmpRs := p.Pop()
				p.CodePointerM = pointerT
				return tk.ToStr(tmpRs)
			}

			p.SetVarInt(pr, deleT)
		case "image.Point", "point":
			// var p1 image.Point
			p.SetVarInt(pr, new(image.Point))

		case "zip":
			vs := p.ParamsToList(instrT, v1p+1)
			tk.Plv(vs)

			p.SetVarInt(pr, archiver.NewZip())

		default:
			return p.ErrStrf("未知对象")
		}

		return ""
	case 1402: // newVar
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		// pr := -5
		// v1p := 0

		// if instrT.ParamLen > 0 {
		pr := instrT.Params[0].Ref
		v1p := 1
		// }

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		switch v1 {
		case "byteList": // 后面可接多个字节，其中可以有字节数组（会逐一加入字节列表中）
			blT := make([]byte, 0)

			vs := p.ParamsToList(instrT, v1p+1)

			for _, vvv := range vs {
				nv, ok := vvv.([]byte)

				if ok {
					for _, vvvj := range nv {
						blT = append(blT, vvvj)
					}
				} else {
					blT = append(blT, tk.ToByte(vvv, 0))
				}
			}

			p.SetVarInt(pr, blT)
		case "time":
			timeT := time.Now()
			p.SetVarInt(pr, &timeT)
		case "gui":
			objT := p.GetVar("guiG")
			p.SetVarInt(pr, objT)
		case "quickDelegate":
			if instrT.ParamLen < 2 {
				return p.ErrStrf("参数不够")
			}

			v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))

			var deleT tk.QuickDelegate

			// same as fastCall
			deleT = func(strA string) string {
				pointerT := p.CodePointerM

				p.Push(strA)

				tmpPointerT := v2

				for {
					rs := p.RunLine(tmpPointerT)

					nv, ok := rs.(int)

					if ok {
						tmpPointerT = nv
						continue
					}

					nsv, ok := rs.(string)

					if ok {
						if tk.IsErrStr(nsv) {
							// tmpRs := p.Pop()
							p.CodePointerM = pointerT
							return nsv
						}

						if nsv == "exit" { // 不应发生
							tmpRs := p.Pop()
							p.CodePointerM = pointerT
							return tk.ToStr(tmpRs)
						} else if nsv == "fr" {
							break
						}
					}

					tmpPointerT++
				}

				// return pointerT + 1

				tmpRs := p.Pop()
				p.CodePointerM = pointerT
				return tk.ToStr(tmpRs)
			}

			p.SetVarInt(pr, deleT)
		case "image.Point", "point":
			var p1 image.Point
			p.SetVarInt(pr, p1)
		// case "strings.Builder", "stringBuffer", "strBuf":
		// 	var p1 strings.Builder

		// 	if instrT.ParamLen > 2 {
		// 		p1.WriteString(tk.ToStr(p.GetVarValue(instrT.Params[2])))
		// 	}

		// 	p.SetVarInt(pr, p1)

		default:
			return p.ErrStrf("未知对象")
		}

		return ""
	case 1403: // method/mt
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v1 := p.GetVarValue(instrT.Params[1])

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[2]))

		switch nv := v1.(type) {
		case string:
			mapT := memberMapG["string"]

			funcT, ok := mapT[v2]

			if ok {
				rsT := callGoFunc(funcT, nv, p.ParamsToList(instrT, 3)...)
				p.SetVarInt(pr, rsT)

				return ""
			}
		case time.Time:
			switch v2 {
			case "toStr":
				p.SetVarInt(pr, fmt.Sprintf("%v", nv))
				return ""
			case "toTick":
				p.SetVarInt(pr, tk.GetTimeStampMid(nv))
				return ""
			case "getInfo":
				zoneT, offsetT := nv.Zone()

				p.SetVarInt(pr, map[string]interface{}{"Time": nv, "Formal": nv.Format(tk.TimeFormat), "Compact": nv.Format(tk.TimeFormat), "Full": fmt.Sprintf("%v", nv), "Year": nv.Year(), "Month": nv.Month(), "Day": nv.Day(), "Hour": nv.Hour(), "Minute": nv.Minute(), "Second": nv.Second(), "Zone": zoneT, "Offset": offsetT, "UnixNano": nv.UnixNano()})
				return ""
			case "format":
				var v2 string = ""

				if instrT.ParamLen > 3 {
					v2 = tk.ToStr(p.GetVarValue(instrT.Params[3]))
				}

				p.SetVarInt(pr, tk.FormatTime(nv, v2))

				return ""
			case "toLocal":
				p.SetVarInt(pr, nv.Local())
				return ""
			case "toGlobal", "toUTC":
				p.SetVarInt(pr, nv.UTC())
				return ""
			case "addDate":
				if instrT.ParamLen < 6 {
					return p.ErrStrf("参数不够")
				}

				v1p := 2

				v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))
				v3 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+2]))
				v4 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+3]))

				p.SetVarInt(pr, nv.AddDate(v2, v3, v4))

				return ""
			case "sub":
				if instrT.ParamLen < 4 {
					return p.ErrStrf("参数不够")
				}

				v1p := 2

				v2 := tk.ToTime(p.GetVarValue(instrT.Params[v1p+1]))

				vvv := nv.Sub(v2.(time.Time))

				p.SetVarInt(pr, int(vvv)/1000000)
				return ""
			default:
				p.SetVarInt(pr, fmt.Sprintf("未知方法: %v", v2))
				return p.ErrStrf("未知方法: %v", v2)
			}
		case *sync.RWMutex:
			switch v2 {
			case "lock":
				nv.Lock()
			case "tryLock":
				nv.TryLock()
			case "readLock":
				nv.RLock()
			case "tryReadLock":
				nv.TryRLock()
			case "unlock":
				nv.Unlock()
			case "readUnlock":
				nv.RUnlock()
			default:
				p.SetVarInt(pr, fmt.Sprintf("未知方法: %v", v2))
				return p.ErrStrf("未知方法: %v", v2)
			}

			p.SetVarInt(pr, "")
			return ""
		case *goph.Client:
			switch v2 {
			case "close":
				nv.Close()
			case "run":
				if instrT.ParamLen < 4 {
					return p.ErrStrf("参数不够")
				}

				v1p := 2

				v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

				rs, errT := nv.Run(v2)

				if errT != nil {
					p.SetVarInt(pr, errT)
				}

				p.SetVarInt(pr, string(rs))

				return ""
			case "upload":
				if instrT.ParamLen < 5 {
					return p.ErrStrf("参数不够")
				}

				v1p := 2

				v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
				v3 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))

				p.SetVarInt(pr, nv.Upload(v2, v3))

				return ""
			case "download":
				if instrT.ParamLen < 5 {
					return p.ErrStrf("参数不够")
				}

				v1p := 2

				v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
				v3 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))

				p.SetVarInt(pr, nv.Download(v2, v3))

				return ""
			case "getFileContent":
				if instrT.ParamLen < 4 {
					return p.ErrStrf("参数不够")
				}

				v1p := 2

				v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

				rs, errT := nv.GetFileContent(v2)

				if errT != nil {
					p.SetVarInt(pr, errT)
				}

				p.SetVarInt(pr, rs)

				return ""
			default:
				p.SetVarInt(pr, fmt.Sprintf("未知方法: %v", v2))
				return p.ErrStrf("未知方法: %v", v2)
			}

			p.SetVarInt(pr, "")
			return ""
		case *strings.Builder:

			switch v2 {
			case "write", "append", "writeString":
				v3 := tk.ToStr(p.GetVarValue(instrT.Params[3]))
				c, errT := nv.WriteString(v3)
				if errT != nil {
					p.SetVarInt(pr, errT)
					return ""
				}

				p.SetVarInt(pr, c)
				return ""
			case "len":
				p.SetVarInt(pr, nv.Len())
				return ""
			case "reset":
				nv.Reset()
				return ""
			case "string", "str", "getStr", "getString":
				p.SetVarInt(pr, nv.String())
				return ""
			default:
				break
				// p.SetVarInt(pr, fmt.Sprintf("未知方法: %v", v2))
				// return p.ErrStrf("未知方法: %v", v2)
			}
			break
			// return ""
		case tk.TXDelegate:
			// if instrT.ParamLen < 3 {
			// 	return p.ErrStrf("参数不够")
			// }

			v1p := 2

			v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
			vs := p.ParamsToList(instrT, v1p+1)

			rs := nv(v2, p, nil, vs...)

			p.SetVarInt(pr, rs)

			return ""
		case *http.Request:
			switch v2 {
			case "saveFormFile":
				if instrT.ParamLen < 6 {
					return p.ErrStrf("参数不够")
				}

				v1p := 2

				v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
				v3 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))
				v4 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+3]))

				argsT := p.ParamsToStrs(instrT, v1p+4)

				formFile1, headerT, errT := nv.FormFile(v2)
				if errT != nil {
					p.SetVarInt(pr, fmt.Sprintf("获取上传文件失败：%v", errT))
					return ""
				}

				defer formFile1.Close()
				tk.Pl("file name : %#v", headerT.Filename)

				defaultExtT := p.GetSwitchVarValue(argsT, "-defaultExt=", "")

				baseT := tk.RemoveFileExt(filepath.Base(headerT.Filename))
				extT := filepath.Ext(headerT.Filename)

				if extT == "" {
					extT = defaultExtT
				}

				v4 = strings.Replace(v4, "TX_fileName_XT", baseT, -1)
				v4 = strings.Replace(v4, "TX_fileExt_XT", extT, -1)

				destFile1, errT := os.CreateTemp(v3, v4) //"pic*.png")
				if errT != nil {
					p.SetVarInt(pr, fmt.Sprintf("保存上传文件失败：%v", errT))
					return ""
				}

				defer destFile1.Close()

				_, errT = io.Copy(destFile1, formFile1)
				if errT != nil {
					p.SetVarInt(pr, fmt.Sprintf("服务器内部错误：%v", errT))
					return ""
				}

				p.SetVarInt(pr, tk.GetLastComponentOfFilePath(destFile1.Name()))
				return ""

			}
			return ""
		}

		mapT := memberMapG[""]

		funcT, ok := mapT[v2]

		if ok {
			rsT := callGoFunc(funcT, v1, p.ParamsToList(instrT, 3)...)
			p.SetVarInt(pr, rsT)

			return ""
		}

		// madarin
		// rv1 := reflect.ValueOf(v1)

		// rv2 := rv1.MethodByName(v2)

		// rvAryT := p.ParamsToReflectValueList(instrT, 3)

		// rrvT := rv2.Call(rvAryT)

		// rvr := make([]interface{}, 0)

		// for _, v9 := range rrvT {
		// 	rvr = append(rvr, v9.Interface{})
		// }
		rvr := tk.ReflectCallMethod(v1, v2, p.ParamsToList(instrT, 3)...)

		p.SetVarInt(pr, rvr)

		// p.SetVarInt(pr, fmt.Errorf("未知方法：（%v）%v", v1, v2))

		return ""
	// return p.ErrStrf("未知方法：（%v）%v", v1, v2)
	case 1405: // member/mb
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v1 := p.GetVarValue(instrT.Params[1])

		// v2 := tk.ToStr(p.GetVarValue(instrT.Params[2]))

		vs := p.ParamsToStrs(instrT, 2)

		var vr interface{} = v1

		for _, v2 := range vs {
			switch nv := vr.(type) {
			case *http.Request:
				switch v2 {
				case "Method":
					vr = nv.Method
					continue
				case "Proto":
					vr = nv.Proto
					continue
				case "Host":
					vr = nv.Host
					continue
				case "RemoteAddr":
					vr = nv.RemoteAddr
					continue
				case "RequestURI":
					vr = nv.RequestURI
					continue
				case "TLS":
					vr = nv.TLS
					continue
				case "URL":
					vr = nv.URL
					continue
				case "Scheme":
					vr = nv.URL.Scheme
					continue
				}

				// p.SetVarInt(pr, fmt.Sprintf("未知成员: %v", v2))
				// return p.ErrStrf("未知成员: %v", v2)

			case *url.URL:
				switch v2 {
				case "Scheme":
					vr = nv.Scheme
					continue
				}

				// p.SetVarInt(pr, fmt.Sprintf("未知成员: %v", v2))
				// return p.ErrStrf("未知成员: %v", v2)

			default:
				break
			}

			typeT := reflect.TypeOf(vr)

			kindT := typeT.Kind()

			// tk.Pl("kind: %v", kindT)

			if kindT == reflect.Ptr {
				vrf := reflect.ValueOf(vr).Elem()

				kindT = vrf.Kind()

				// tk.Pl("vrf: %v, kind: %v", vrf, kindT)

				if kindT == reflect.Struct {
					vr = vrf.Interface()
				}
			}

			if kindT == reflect.Struct {
				rv1 := reflect.ValueOf(vr)

				rv2 := rv1.FieldByName(v2).Interface()

				vr = rv2
				continue
			}

			return p.ErrStrf("未知成员：%v（%T/%v）.%v", vr, vr, kindT, v2)

		}

		p.SetVarInt(pr, vr)
		return ""
	case 1407: // mbSet
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		v3 := p.GetVarValue(instrT.Params[v1p+2])

		// var vr interface{} = v1

		switch nv := v1.(type) {
		case string:
			return p.ErrStrf("无法处理的类型：%v（%T/%v）", nv, v1, v1)
		default:
			typeT := reflect.TypeOf(v1)

			kindT := typeT.Kind()

			// tk.Pl("kind: %v", kindT)
			if kindT == reflect.Ptr {
				rv1 := reflect.ValueOf(v1)

				// tk.Pl("v1: %#v, p-rv1: %v, canSet: %v", v1, rv1, rv1.CanSet())
				vrf := rv1.Elem()
				// tk.Pl("vrf: %v, canSet: %v", vrf, vrf.CanSet())

				kindT = vrf.Kind()

				// tk.Pl("vrf: %v, kind: %v", vrf, kindT)

				if kindT == reflect.Struct {
					rv2 := vrf.FieldByName(v2)

					rv3 := reflect.ValueOf(v3)
					// tk.Pl("rv3: %v", rv3)

					rv2.Set(rv3)

					break

				}
			} else if kindT == reflect.Struct {
				rv1 := reflect.ValueOf(v1)
				// tk.Pl("v1: %#v, rv1: %#v, canSet: %v", v1, rv1, rv1.CanSet())

				rv2 := rv1.FieldByName(v2)

				rv3 := reflect.ValueOf(v3)
				// tk.Pl("rv3: %v", rv3)

				rv2.Set(rv3)

				break
			}

			return p.ErrStrf("无法处理的类型：%v（%T/%v）", v1, v1, kindT)

		}

		p.SetVarInt(pr, "")
		return ""
	case 1410: // newObj
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

		switch v1 {
		case "string":
			objT := &XieString{}

			objT.Init(p.ParamsToList(instrT, 2)...)

			p.SetVarInt(pr, objT)
		case "any":
			objT := &XieAny{}

			objT.Init(p.ParamsToList(instrT, 2)...)

			p.SetVarInt(pr, objT)
		case "mux":
			p.SetVarInt(pr, http.NewServeMux())
		case "int":
			p.SetVarInt(pr, new(int))
		case "byte":
			p.SetVarInt(pr, new(byte))
		default:
			return p.ErrStrf("未知对象")
		}

		return ""
	case 1411: // setObjValue
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(XieObject)

		// v2 := p.GetVarValue(instrT.Params[0])

		v1.SetValue(p.ParamsToList(instrT, 1)...)

		return ""
	case 1412: // getObjValue
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p]).(XieObject)

		p.SetVarInt(pr, v1.GetValue())

		return ""
	case 1440: // callObj
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(XieObject)

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

		argsT := []interface{}{v2}

		argsT = append(argsT, p.ParamsToList(instrT, 2)...)

		rsT := v1.Call(argsT...)

		if tk.IsError(rsT) {
			return p.ErrStrf("对象方法调用失败：%v", rsT)
		}

		if rsT != nil {
			p.Push(rsT)
		}

		return ""
	case 1501: // backQuote
		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		p.SetVarInt(pr, "`")

		return ""
	case 1503: // quote
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rs := strconv.Quote(v1)

		p.SetVarInt(pr, rs[1:len(rs)-1])

		return ""
	case 1504: // unquote
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rs, errT := strconv.Unquote(`"` + v1 + `"`)

		if errT != nil {
			p.ErrStrf("unquote失败：%v", errT)
		}

		p.SetVarInt(pr, rs)

		return ""
	case 1510: // isEmpty
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, (v1 == ""))

		return ""

	case 1520: // strAdd
		pr := -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Pop()
			v1 = p.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0].Ref
			v2 = p.Pop()
			v1 = p.Pop()
			// return p.ErrStrf("参数不够")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(instrT.Params[0])
			v2 = p.GetVarValue(instrT.Params[1])
		} else {
			pr = instrT.Params[0].Ref
			v1 = p.GetVarValue(instrT.Params[1])
			v2 = p.GetVarValue(instrT.Params[2])
		}

		p.SetVarInt(pr, v1.(string)+v2.(string))

		return ""

	case 1530: // strSplit
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		s1 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

		s2 := tk.ToStr(p.GetVarValue(instrT.Params[2]))

		countT := -1

		if instrT.ParamLen > 3 {
			countT = tk.ToInt(p.GetVarValue(instrT.Params[3]))
		}

		listT := strings.SplitN(s1, s2, countT)

		p.SetVarInt(pr, listT)

		return ""

	case 1540: // strReplace
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		v3 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))

		p.SetVarInt(pr, strings.ReplaceAll(v1, v2, v3))

		return ""

	case 1543: // strReplaceIn
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		vs := p.ParamsToStrs(instrT, v1p+1)

		p.SetVarInt(pr, tk.StringReplace(v1, vs...))

		return ""

	case 1550: // trim/strTrim
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		if v1 == nil {
			p.SetVarInt(pr, "")
		} else if v1 == Undefined {
			p.SetVarInt(pr, "")
		} else {
			p.SetVarInt(pr, strings.TrimSpace(tk.ToStr(v1)))
		}

		return ""

	case 1551: // trimSet
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		p.SetVarInt(pr, strings.Trim(v1, v2))

		return ""

	case 1553: // trimSetLeft
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		p.SetVarInt(pr, strings.TrimLeft(v1, v2))

		return ""

	case 1554: // trimSetRight
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		p.SetVarInt(pr, strings.TrimRight(v1, v2))

		return ""

	case 1557: // trimPrefix
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		p.SetVarInt(pr, strings.TrimPrefix(v1, v2))

		return ""

	case 1558: // trimSuffix
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		p.SetVarInt(pr, strings.TrimSuffix(v1, v2))

		return ""

	case 1561: // toUpper
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, strings.ToUpper(v1))

		return ""

	case 1562: // toLower
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, strings.ToLower(v1))

		return ""

	case 1563: // strPad
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		p1 := instrT.Params[0].Ref

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))
		v3 := tk.ToInt(p.GetVarValue(instrT.Params[2]))

		p.SetVarInt(p1, tk.PadString(v2, v3, p.ParamsToStrs(instrT, 2)...))

		return ""

	case 1571: // strContains
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		p.SetVarInt(pr, strings.Contains(v1, v2))

		return ""

	case 1572: // strContainsIn
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := p.ParamsToStrs(instrT, v1p+1)

		p.SetVarInt(pr, tk.ContainsIn(v1, v2...))

		return ""

	case 1573: // strCount
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		p.SetVarInt(pr, strings.Count(v1, v2))

		return ""

	case 1581: // strIn/inStrs
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

		v3i, ok := p.GetVarValue(instrT.Params[2]).([]string)

		if ok {
			p.SetVarInt(pr, tk.InStrings(v2, v3i...))
			return ""
		}

		v3 := p.ParamsToStrs(instrT, 2)

		p.SetVarInt(pr, tk.InStrings(v2, v3...))

		return ""

	// case 1582: // strInArray
	// 	if instrT.ParamLen < 3 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	p1 := instrT.Params[0].Ref

	// 	v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

	// 	v3 := p.GetVarValue(instrT.Params[2]).([]string)

	// 	p.SetVarInt(p1, tk.InStrings(v2, v3...))

	// 	return ""

	case 1582: // strStartsWith
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		p.SetVarInt(pr, strings.HasPrefix(v1, v2))

		return ""

	case 1583: // strStartsWithIn
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := p.ParamsToStrs(instrT, v1p+1)

		p.SetVarInt(pr, tk.StartsWith(v1, v2...))

		return ""

	case 1584: // strEndsWith
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		p.SetVarInt(pr, strings.HasSuffix(v1, v2))

		return ""

	case 1585: // strEndsWithIn
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := p.ParamsToStrs(instrT, v1p+1)

		p.SetVarInt(pr, tk.EndsWith(v1, v2...))

		return ""

	case 1601: // bytesToData
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := p.GetVarValue(instrT.Params[v1p]).([]byte)
		v2 := p.GetVarValue(instrT.Params[v1p+1])

		vs := p.ParamsToStrs(instrT, v1p+2)

		p.SetVarInt(pr, tk.BytesToData(v1, v2, vs...))

		return ""

	case 1603: // dataToBytes
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := p.GetVarValue(instrT.Params[v1p])

		vs := p.ParamsToStrs(instrT, v1p+1)

		p.SetVarInt(pr, tk.DataToBytes(v1, vs...))

		return ""

	case 1605: // bytesToHex
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := p.GetVarValue(instrT.Params[v1p]).([]byte)

		p.SetVarInt(pr, tk.BytesToHex(v1))

		return ""

	case 1910: // now
		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		errT := p.SetVarInt(pr, time.Now())

		if errT != nil {
			return p.ErrStrf("%v", errT)
		}

		return ""

	case 1911: // nowStrCompact
		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		errT := p.SetVarInt(pr, tk.GetNowTimeString())

		if errT != nil {
			return p.ErrStrf("%v", errT)
		}

		return ""

	case 1912: // nowStr/nowStrFormal
		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		errT := p.SetVarInt(pr, tk.GetNowTimeStringFormal())

		if errT != nil {
			return p.ErrStrf("%v", errT)
		}

		return ""
	case 1913: // nowTick
		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		errT := p.SetVarInt(pr, tk.GetTimeStampMid(time.Now()))

		if errT != nil {
			return p.ErrStrf("%v", errT)
		}

		return ""
	case 1918: // nowUTC
		pr := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		errT := p.SetVarInt(pr, time.Now().UTC())

		if errT != nil {
			return p.ErrStrf("%v", errT)
		}

		return ""

	case 1921: // timeSub
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		sd := int(v1.(time.Time).Sub(v2.(time.Time)))

		p.SetVarInt(pr, sd/1000000)

		return ""

	case 1941: // timeToLocal
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToTime(p.GetVarValue(instrT.Params[v1p]))

		if tk.IsError(v1) {
			return p.ErrStrf("时间转换失败：%v", v1)
		}

		p.SetVarInt(pr, v1.(time.Time).Local())

		return ""

	case 1942: // timeToGlobal
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToTime(p.GetVarValue(instrT.Params[v1p]))

		if tk.IsError(v1) {
			return p.ErrStrf("时间转换失败：%v", v1)
		}

		p.SetVarInt(pr, v1.(time.Time).UTC())

		return ""

	case 1951: // getTimeInfo
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToTime(p.GetVarValue(instrT.Params[v1p]))

		if tk.IsError(v1) {
			return p.ErrStrf("时间转换失败：%v", v1)
		}

		nv := v1.(time.Time)

		zoneT, offsetT := nv.Zone()

		p.SetVarInt(pr, map[string]interface{}{"Time": nv, "Formal": nv.Format(tk.TimeFormat), "Compact": nv.Format(tk.TimeFormat), "Full": fmt.Sprintf("%v", nv), "Year": nv.Year(), "Month": nv.Month(), "Day": nv.Day(), "Hour": nv.Hour(), "Minute": nv.Minute(), "Second": nv.Second(), "Zone": zoneT, "Offset": offsetT})

		return ""

	case 1961: // timeAddDate
		if instrT.ParamLen < 5 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToTime(p.GetVarValue(instrT.Params[v1p]))

		if tk.IsError(v1) {
			return p.ErrStrf("时间转换失败：%v", v1)
		}

		v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))
		v3 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+2]))
		v4 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+3]))

		nv := v1.(time.Time)

		p.SetVarInt(pr, nv.AddDate(v2, v3, v4))

		return ""

	case 1971: // formatTime
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToTime(p.GetVarValue(instrT.Params[v1p]))

		if tk.IsError(v1) {
			return p.ErrStrf("时间转换失败：%v", v1)
		}

		var v2 string = ""

		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
		}

		p.SetVarInt(pr, tk.FormatTime(v1.(time.Time), v2))

		return ""

	// case 1981: // toTime
	// 	if instrT.ParamLen < 2 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	pr := instrT.Params[0].Ref
	// 	v1p := 1

	// 	v1 := tk.ToTime(p.GetVarValue(instrT.Params[v1p]), p.ParamsToList(instrT, v1p+1)...)

	// 	p.SetVarInt(pr, v1)

	// 	return ""

	case 1991: // timeToTick
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToTime(p.GetVarValue(instrT.Params[v1p]), p.ParamsToList(instrT, v1p+1)...)

		if tk.IsErrX(v1) {
			p.SetVarInt(pr, v1)
			return ""
		}

		t := tk.GetTimeStampMid(v1.(time.Time))

		p.SetVarInt(pr, t)

		return ""

	case 1993: // tickToTime
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToInt(p.GetVarValue(instrT.Params[v1p]))

		if v1 == -1 {
			p.SetVarInt(pr, fmt.Errorf("转换时间戳失败：%v", p.GetVarValue(instrT.Params[v1p])))
			return ""
		}

		t := time.Unix(int64(v1), 0)

		p.SetVarInt(pr, t)

		return ""

	case 2100: // abs
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		switch nv := v1.(type) {
		case int:
			p.SetVarInt(pr, tk.AbsInt(nv))
		case float64:
			if nv < 0 {
				p.SetVarInt(pr, -nv)
			} else {
				p.SetVarInt(pr, nv)
			}
		case byte:
			if nv < 0 {
				p.SetVarInt(pr, -nv)
			} else {
				p.SetVarInt(pr, nv)
			}
		case rune:
			if nv < 0 {
				p.SetVarInt(pr, -nv)
			} else {
				p.SetVarInt(pr, nv)
			}
		default:
			return p.ErrStrf("无法处理的类型：%T", v1)
		}

		return ""

	case 10001: // getParam
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		v3 := p.GetVarValue(instrT.Params[v1p+2])

		v1n, ok := v1.([]string)

		if ok {
			p.SetVarInt(pr, tk.GetParameterByIndexWithDefaultValue(v1n, tk.ToInt(v2), tk.ToStr(v3)))
			return ""
		}

		v2n, ok := v1.([]interface{})

		if ok {
			p.SetVarInt(pr, tk.GetParamI(v2n, tk.ToInt(v2), tk.ToStr(v3)))
			return ""
		}

		return p.ErrStrf("参数类型错误")

	case 10002: // getSwitch
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		v1p := 0
		pr := -5

		if instrT.ParamLen > 3 {
			v1p = 1
			pr = instrT.Params[0].Ref
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		defaultT := tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))

		v1n, ok := v1.([]string)

		if ok {
			p.SetVarInt(pr, tk.GetSwitch(v1n, v2, defaultT))
			return ""
		}

		v2n, ok := v1.([]interface{})

		if ok {
			p.SetVarInt(pr, tk.GetSwitchI(v2n, v2, defaultT))
			return ""
		}

		return p.ErrStrf("参数类型错误")

	case 10003: // ifSwitchExists
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1p := 0
		pr := -5

		if instrT.ParamLen > 2 {
			v1p = 1
			pr = instrT.Params[0].Ref
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		v1n, ok := v1.([]string)

		if ok {
			p.SetVarInt(pr, tk.IfSwitchExistsWhole(v1n, tk.ToStr(v2)))
			return ""
		}

		v2n, ok := v1.([]interface{})

		if ok {
			p.SetVarInt(pr, tk.IfSwitchExistsWholeI(v2n, tk.ToStr(v2)))
			return ""
		}

		return p.ErrStrf("参数类型错误")

	case 10005: // ifSwitchNotExists/switchNotExists
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1p := 0
		pr := -5

		if instrT.ParamLen > 2 {
			v1p = 1
			pr = instrT.Params[0].Ref
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		v1n, ok := v1.([]string)

		if ok {
			p.SetVarInt(pr, !tk.IfSwitchExistsWhole(v1n, tk.ToStr(v2)))
			return ""
		}

		v2n, ok := v1.([]interface{})

		if ok {
			p.SetVarInt(pr, !tk.IfSwitchExistsWholeI(v2n, tk.ToStr(v2)))
			return ""
		}

		return p.ErrStrf("参数类型错误")

	case 10410: // pln
		list1T := []interface{}{}

		for _, v := range instrT.Params {
			list1T = append(list1T, p.GetVarValue(v))
		}

		fmt.Println(list1T...)

		return ""
	case 10411: // plo
		if instrT.ParamLen < 1 {
			tk.Plo(p.TmpM)
			return ""
		}

		valueT := p.GetVarValue(instrT.Params[0])

		tk.Plo(valueT)

		return ""
	case 10412: // plos
		if instrT.ParamLen < 1 {
			tk.Plos(p.TmpM)
			return ""
		}

		vs := p.ParamsToList(instrT, 0)

		tk.Plos(vs...)

		return ""
	case 10420: // pl
		list1T := []interface{}{}

		formatT := ""

		for i, v := range instrT.Params {
			if i == 0 {
				formatT = v.Value.(string)
				continue
			}

			list1T = append(list1T, p.GetVarValue(v))
		}

		fmt.Printf(formatT+"\n", list1T...)

		return ""
	case 10422: // plNow
		list1T := []interface{}{}

		formatT := ""

		for i, v := range instrT.Params {
			if i == 0 {
				formatT = v.Value.(string)
				continue
			}

			list1T = append(list1T, p.GetVarValue(v))
		}

		tk.PlNow(formatT, list1T...)

		return ""
	case 10430: // plv
		if instrT.ParamLen < 1 {
			tk.Plv(p.TmpM)
			return ""
			// return p.ErrStrf("参数不够")
		}

		s1 := p.GetVarValue(instrT.Params[0])

		tk.Plv(s1)

		return ""

	case 10433: // plvsr

		vs := p.ParamsToList(instrT, 0)

		tk.Plvsr(vs)

		return ""

	case 10440: // plErr
		if instrT.ParamLen < 1 {
			tk.PlErr(p.TmpM.(error))
			return ""
			// return p.ErrStrf("参数不够")
		}

		s1 := p.GetVarValue(instrT.Params[0]).(error)

		tk.PlErr(s1)

		return ""

	case 10450: // plErrStr
		if instrT.ParamLen < 1 {
			tk.PlErrString(p.TmpM.(string))
			return ""
			// return p.ErrStrf("参数不够")
		}

		s1 := p.GetVarValue(instrT.Params[0]).(string)

		tk.PlErrString(s1)

		return ""

	case 10460: // spr
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

		v3 := p.ParamsToList(instrT, 2)

		errT := p.SetVarInt(pr, fmt.Sprintf(v2, v3...))

		if errT != nil {
			return p.ErrStrf("变量赋值错误：%v", errT)
		}

		return ""
	case 10511: // scanf
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[0]))

		vs := p.ParamsToList(instrT, 1)

		_, errT := fmt.Scanf(v1, vs...)

		if errT != nil {
			return p.ErrStrf("扫描数据失败：%v", errT)
		}

		return ""
	case 10512: // sscanf
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[0]))

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

		v3 := p.ParamsToList(instrT, 2)

		_, errT := fmt.Sscanf(v1, v2, v3...)

		if errT != nil {
			return p.ErrStrf("扫描数据失败：%v", errT)
		}

		return ""
	case 10810: // convert/转换
		// tk.Plv(instrT)
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数个数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		if tk.IsError(v1) {
			return p.ErrStrf("参数错误")
		}

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		if tk.IsError(v2) {
			return p.ErrStrf("参数错误")
		}

		var v3 interface{}

		s2 := v2.(string)

		if s2 == "bool" || s2 == "布尔" {
			v3 = tk.ToBool(v1)
		} else if s2 == "int" || s2 == "整数" {
			v3 = tk.ToInt(v1)
		} else if s2 == "byte" || s2 == "字节" {
			v3 = tk.ToByte(v1)
		} else if s2 == "rune" || s2 == "如痕" {
			v3 = tk.ToRune(v1)
		} else if s2 == "float" || s2 == "小数" {
			v3 = tk.ToFloat(v1)
		} else if s2 == "str" || s2 == "字符串" {
			// nv, ok := v1.(XieObject)

			// if ok {
			// 	v3 = nv.Call("toStr")
			// } else {
			v3 = tk.ToStr(v1)
			// }
		} else if s2 == "list" || s2 == "列表" {

			switch nv := v1.(type) {
			case []interface{}:
				v3 = nv
			case []string:
				listT := make([]interface{}, 0, len(nv))

				for _, v := range nv {
					listT = append(listT, v)
				}

				v3 = listT
			default:
				return p.ErrStrf("无法处理的类型")
			}
		} else if s2 == "strList" || s2 == "字符串列表" {

			switch nv := v1.(type) {
			case []string:
				v3 = nv
			case []interface{}:
				listT := make([]string, 0, len(nv))

				for _, v := range nv {
					listT = append(listT, tk.ToStr(v))
				}

				v3 = listT
			case []VarRef:
				listT := make([]string, 0, len(nv))

				for _, v := range nv {
					listT = append(listT, tk.ToStr(v.Value))
				}

				v3 = listT
			default:
				return p.ErrStrf("无法处理的类型")
			}
		} else if s2 == "byteList" || s2 == "字节列表" {

			switch nv := v1.(type) {
			case string:
				v3 = []byte(nv)
			case []rune:
				v3 = []byte(string(nv))
			default:
				return p.ErrStrf("无法处理的类型")
			}
		} else if s2 == "runeList" || s2 == "如痕列表" {

			switch nv := v1.(type) {
			case string:
				v3 = []rune(nv)
			case []byte:
				v3 = []byte(string(nv))
			default:
				return p.ErrStrf("无法处理的类型")
			}
		} else if s2 == "map" || s2 == "映射" {

			switch nv := v1.(type) {
			case map[string]interface{}:
				v3 = nv
			case map[string]string:
				mapT := make(map[string]interface{}, len(nv))

				for k, v := range nv {
					mapT[k] = v
				}

				v3 = mapT
			default:
				return p.ErrStrf("无法处理的类型")
			}
		} else if s2 == "strMap" || s2 == "字符串映射" {

			switch nv := v1.(type) {
			case map[string]string:
				v3 = nv
			case map[string]interface{}:
				mapT := make(map[string]string, len(nv))

				for k, v := range nv {
					mapT[k] = tk.ToStr(v)
				}

				v3 = mapT
			default:
				return p.ErrStrf("无法处理的类型")
			}
		} else if s2 == "time" || s2 == "时间" {

			switch nv := v1.(type) {
			case time.Time:
				v3 = nv
			case string:
				v3 = tk.ToTime(nv, p.ParamsToList(instrT, v1p+2)...)
			default:
				tmps := tk.ToStr(v1)
				v3 = tk.ToTime(tmps, p.ParamsToList(instrT, v1p+2)...)
			}
		} else if s2 == "timeStr" || s2 == "时间字符串" {

			switch nv := v1.(type) {
			case time.Time:
				v3 = tk.FormatTime(nv, p.ParamsToStrs(instrT, v1p+2)...)
			case string:
				rs := tk.ToTime(nv)
				if tk.IsError(rs) {
					return p.ErrStrf("时间转换失败：%v", rs)
				}

				v3 = tk.FormatTime(rs.(time.Time), p.ParamsToStrs(instrT, v1p+2)...)
			default:
				rs := tk.ToTime(tk.ToStr(v1))
				if tk.IsError(rs) {
					return p.ErrStrf("时间转换失败：%v", rs)
				}

				v3 = tk.FormatTime(rs.(time.Time), p.ParamsToStrs(instrT, v1p+2)...)
			}
		} else if s2 == "tick" || s2 == "timeStamp" || s2 == "时间戳" {

			switch nv := v1.(type) {
			case time.Time:
				v3 = tk.GetTimeStampMid(nv)
			default:
				p.ErrStrf("类型不匹配：%v", v1)
			}
		} else {
			return p.ErrStrf("无法处理的类型")
		}

		p.SetVarInt(pr, v3)

		return ""

	// case 10811: // convert$
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	s1 := p.Pop()

	// 	v2 := p.GetVarValue(instrT.Params[1])

	// 	if tk.IsError(v2) {
	// 		return p.ErrStrf("参数错误")
	// 	}

	// 	s2 := v2.(string)

	// 	if s2 == "b" {
	// 		p.Push(tk.ToBool(s1))
	// 	} else if s2 == "i" {
	// 		p.Push(tk.ToInt(s1))
	// 	} else if s2 == "f" {
	// 		p.Push(tk.ToFloat(s1))
	// 	} else if s2 == "s" {
	// 		p.Push(tk.ToStr(s1))
	// 	} else {
	// 		return p.ErrStrf("unknown type")
	// 	}

	// 	return ""

	case 10821: // hex
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		var v3 interface{}

		v3 = tk.DataToBytes(v1, "-endian=B")

		if tk.IsError(v3) {
			return p.ErrStrf("转换失败：%v", v3)
		}

		v3 = hex.EncodeToString(v3.([]byte))
		p.SetVarInt(pr, v3)

		return ""

	case 10822: // hexb
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		var v3 interface{}

		v3 = tk.DataToBytes(v1, "-endian=L")

		if tk.IsError(v3) {
			return p.ErrStrf("转换失败：%v", v3)
		}

		v3 = hex.EncodeToString(v3.([]byte))
		p.SetVarInt(pr, v3)

		return ""

	case 10823: // unhex
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, tk.HexToBytes(v1))

		return ""

	case 10824: // toHex
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		p.SetVarInt(pr, tk.ToHex(v1))

		return ""

	case 10831: // toBool
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		p.SetVarInt(pr, tk.ToBool(v1))

		return ""

	case 10835: // toByte
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		p.SetVarInt(pr, tk.ToByte(v1))

		return ""

	case 10837: // toRune
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		p.SetVarInt(pr, tk.ToRune(v1))

		return ""

	case 10851: // toInt
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		if instrT.ParamLen > 2 {
			v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))

			p.SetVarInt(pr, tk.ToInt(v1, v2))

			return ""
		}

		p.SetVarInt(pr, tk.ToInt(v1))

		return ""

	case 10855: // toFloat
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		if instrT.ParamLen > 2 {
			v2 := tk.ToFloat(p.GetVarValue(instrT.Params[v1p+1]))

			p.SetVarInt(pr, tk.ToFloat(v1, v2))

			return ""
		}

		p.SetVarInt(pr, tk.ToFloat(v1))

		return ""

	case 10861: // toStr
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		p.SetVarInt(pr, tk.ToStr(v1))

		return ""

	case 10871: // toTime
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToTime(p.GetVarValue(instrT.Params[v1p]), p.ParamsToList(instrT, v1p+1)...)

		p.SetVarInt(pr, v1)

		return ""
	// if instrT.ParamLen < 1 {
	// 	return p.ErrStrf("参数不够")
	// }

	// pr := -5
	// v1p := 0

	// if instrT.ParamLen > 1 {
	// 	pr = instrT.Params[0].Ref
	// 	v1p = 1
	// }

	// v1 := p.GetVarValue(instrT.Params[v1p])

	// if instrT.ParamLen > 2 {
	// 	v2 := tk.ToTime(p.GetVarValue(instrT.Params[v1p+1]))

	// 	p.SetVarInt(pr, tk.ToTime(v1, v2))

	// 	return ""
	// }

	// p.SetVarInt(pr, tk.ToTime(v1))

	// return ""

	case 10891: // toAny
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		switch nv := v1.(type) {
		case reflect.Value:
			p.SetVarInt(pr, nv.Interface())
			return ""
		}

		p.SetVarInt(pr, interface{}(v1))

		return ""

	case 10910: // isErrStr
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p]).(string)

		var rsT bool

		errMsgT := ""

		if tk.IsErrStr(v1) {
			rsT = true
		} else {
			rsT = false
		}

		p.SetVarInt(pr, rsT)

		if instrT.ParamLen > 2 {
			if rsT {
				errMsgT = tk.GetErrStr(v1)
			}

			p.SetVarInt(instrT.Params[2].Ref, errMsgT)
		}

		return ""

	case 10921: // getErrStr
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, tk.GetErrStr(v1))

		return ""

	// case 10922: // getErrStr$
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("参数不够")
	// 	}

	// 	v1 := p.GetVarValue(instrT.Params[0]).(string)

	// 	p.Push(tk.GetErrStr(v1))

	// 	return ""

	case 10931: // checkErrStr
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[0]))

		if tk.IsErrStr(v1) {
			// tk.Pln(v1)
			return p.ErrStrf(tk.GetErrStr(v1))
			// return "exit"
		}

		return ""

	case 10941: // isErr
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		var rsT bool

		errMsgT := ""

		if tk.IsError(v1) {
			rsT = true
		} else {
			rsT = false
		}

		p.SetVarInt(pr, rsT)

		if instrT.ParamLen > 2 {
			if rsT {
				errMsgT = v1.(error).Error()
			}

			p.SetVarInt(instrT.Params[2].Ref, errMsgT)
		}

		return ""
	case 10942: // getErrMsg
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p]).(error)

		p.SetVarInt(pr, v1.Error())

		return ""

	case 10943: // isErrX
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := p.GetVarValue(instrT.Params[v1p])

		if tk.IsErrX(v1) {
			if instrT.ParamLen > 2 {
				p2 := instrT.Params[v1p+1].Ref
				p.SetVarInt(p2, tk.GetErrStrX(v1))
			}

			p.SetVarInt(pr, true)

			return ""
		}

		if instrT.ParamLen > 2 {
			p2 := instrT.Params[v1p+1].Ref
			p.SetVarInt(p2, "")
		}

		p.SetVarInt(pr, false)

		return ""
	case 10945: // checkErrX
		if instrT.ParamLen < 1 {
			if tk.IsErrX(p.TmpM) {
				return p.ErrStrf(tk.GetErrStrX(p.TmpM))
			}

			return ""
			// return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		if tk.IsErrX(v1) {
			return p.ErrStrf(tk.GetErrStrX(v1))
		}

		return ""
	case 20110: // writeResp
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(http.ResponseWriter)

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

		tk.WriteResponse(v1, v2)

		return ""

	case 20111: // setRespHeader
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(http.ResponseWriter)

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

		v3 := tk.ToStr(p.GetVarValue(instrT.Params[2]))

		v1.Header().Set(v2, v3)

		return ""

	case 20112: // writeRespHeader
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(http.ResponseWriter)

		v2 := tk.ToInt(p.GetVarValue(instrT.Params[1]))

		v1.WriteHeader(v2)

		return ""

	case 20113: // getReqHeader
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p]).(*http.Request)

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		p.SetVarInt(pr, v1.Header.Get(v2))

		return ""

	case 20114: // genJsonResp/genResp
		if instrT.ParamLen < 4 {
			return p.ErrStrf("参数不够：%v", instrT.ParamLen)
		}

		pr := instrT.Params[0].Ref

		v2 := p.GetVarValue(instrT.Params[1]).(*http.Request)

		v3 := tk.ToStr(p.GetVarValue(instrT.Params[2]))

		v4 := tk.ToStr(p.GetVarValue(instrT.Params[3]))

		rsT := tk.GenerateJSONPResponseWithMore(v3, v4, v2, p.ParamsToStrs(instrT, 4)...)

		p.SetVarInt(pr, rsT)

		return ""

	case 20116: // serveFile
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够：%v", instrT.ParamLen)
		}

		// pr := instrT.Params[0].Ref

		v2 := p.GetVarValue(instrT.Params[0]).(http.ResponseWriter)

		v3 := p.GetVarValue(instrT.Params[1]).(*http.Request)

		v4 := tk.ToStr(p.GetVarValue(instrT.Params[2]))

		http.ServeFile(v2, v3, v4)

		return ""

	case 20121: // newMux
		p1 := -5

		if instrT.ParamLen > 0 {
			p1 = instrT.Params[0].Ref
		}

		p.SetVarInt(p1, http.NewServeMux())

		return ""

	case 20122: // setMuxHandler
		if instrT.ParamLen < 4 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(*http.ServeMux)
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))
		v3 := p.GetVarValue(instrT.Params[2])
		v4 := tk.ToStr(p.GetVarValue(instrT.Params[3]))

		// var inputG interface{}

		// if instrT.ParamLen > 3 {
		// 	inputG = p.GetVarValue(instrT.Params[3])
		// } else {
		// 	inputG = make(map[string]interface{})
		// }

		fnT := func(res http.ResponseWriter, req *http.Request) {
			if res != nil {
				res.Header().Set("Access-Control-Allow-Origin", "*")
				res.Header().Set("Access-Control-Allow-Headers", "*")
				res.Header().Set("Content-Type", "text/html; charset=utf-8")
			}

			if req != nil {
				req.ParseForm()
				req.ParseMultipartForm(1000000000000)
			}

			var paraMapT map[string]string

			paraMapT = tk.FormToMap(req.Form)

			toWriteT := ""

			vmT := NewXie(p.SharedMapM)

			vmT.SetVar("paraMapG", paraMapT)
			vmT.SetVar("requestG", req)
			vmT.SetVar("responseG", res)
			vmT.SetVar("reqNameG", req.RequestURI)
			vmT.SetVar("inputG", v3)

			lrs := vmT.Load(v4)

			if tk.IsErrStr(lrs) {
				res.Write([]byte(fmt.Sprintf("操作失败：%v", tk.GetErrStr(lrs))))
				return
			}

			rs := vmT.Run()

			if tk.IsErrStr(rs) {
				res.Write([]byte(fmt.Sprintf("操作失败：%v", tk.GetErrStr(rs))))
				return
			}

			toWriteT = rs

			if toWriteT == "TX_END_RESPONSE_XT" {
				return
			}

			res.Header().Set("Content-Type", "text/html; charset=utf-8")

			res.Write([]byte(toWriteT))

		}

		v1.HandleFunc(v2, fnT)

		return ""

	case 20123: // setMuxStaticDir
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(*http.ServeMux)
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))
		v3 := tk.ToStr(p.GetVarValue(instrT.Params[2]))
		// v4 := tk.ToStr(p.GetVarValue(instrT.Params[3]))

		// tk.Pln(v1, v2, v3)

		var staticFS http.Handler = http.StripPrefix(v2, http.FileServer(http.Dir(v3)))

		serveStaticDirHandler := func(w http.ResponseWriter, r *http.Request) {
			// if staticFS == nil {
			// tk.Pl("staticFS: %#v", staticFS)
			// staticFS = http.StripPrefix("/w/", http.FileServer(http.Dir(filepath.Join(basePathG, "w"))))
			// hdl :=
			// tk.Pl("hdl: %#v", hdl)
			// staticFS = hdl
			// }
			// tk.Pl("here: %v", r)

			old := r.URL.Path

			// tk.Pl("urlPath: %v", r.URL.Path)
			tk.Pl("old: %v", path.Clean(old))
			// tk.Pl("v2: %v", v2)
			// tk.Pl("trim: %v", strings.TrimPrefix(path.Clean(old), v2))

			name := filepath.Join(v3, strings.TrimPrefix(path.Clean(old), v2))

			tk.Pl("name: %v", name)

			info, err := os.Lstat(name)
			if err == nil {
				if !info.IsDir() {
					staticFS.ServeHTTP(w, r)
					// http.ServeFile(w, r, name)
				} else {
					if tk.IfFileExists(filepath.Join(name, "index.html")) {
						staticFS.ServeHTTP(w, r)
					} else {
						http.NotFound(w, r)
					}
				}
			} else {
				http.NotFound(w, r)
			}

		}

		v1.HandleFunc(v2, serveStaticDirHandler)

		return ""

	case 20151: // startHttpServer
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		p1 := instrT.Params[0].Ref
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))
		if !strings.HasPrefix(v2, ":") {
			v2 = ":" + v2
		}
		v3 := p.GetVarValue(instrT.Params[2]).(*http.ServeMux)

		ifGoT := tk.IfSwitchExists(p.ParamsToStrs(instrT, 3), "-go")

		if ifGoT {
			go http.ListenAndServe(v2, v3)
			p.SetVarInt(p1, "")

			return ""
		}

		errT := http.ListenAndServe(v2, v3)

		if errT != nil {
			p.SetVarInt(p1, fmt.Errorf("启动服务失败：%v", errT))
		} else {
			p.SetVarInt(p1, "")
		}

		return ""

	case 20153: // startHttpsServer
		if instrT.ParamLen < 5 {
			return p.ErrStrf("参数不够")
		}

		p1 := instrT.Params[0].Ref
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[1]))
		if !strings.HasPrefix(v2, ":") {
			v2 = ":" + v2
		}
		v3 := p.GetVarValue(instrT.Params[2]).(*http.ServeMux)
		v4 := tk.ToStr(p.GetVarValue(instrT.Params[3]))
		v5 := tk.ToStr(p.GetVarValue(instrT.Params[4]))

		ifGoT := tk.IfSwitchExists(p.ParamsToStrs(instrT, 5), "-go")

		if ifGoT {
			go http.ListenAndServeTLS(v2, v4, v5, v3)
			p.SetVarInt(p1, "")

			return ""
		}

		errT := http.ListenAndServeTLS(v2, v4, v5, v3)

		if errT != nil {
			p.SetVarInt(p1, fmt.Errorf("启动服务失败：%v", errT))
		} else {
			p.SetVarInt(p1, "")
		}

		return ""

	case 20210: // getWeb
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v2 := p.GetVarValue(instrT.Params[1])

		listT := p.ParamsToList(instrT, 2)

		rs := tk.DownloadWebPageX(tk.ToStr(v2), listT...)

		p.SetVarInt(pr, rs)

		return ""

	case 20220: // downloadFile
		if instrT.ParamLen < 4 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v1p := 1

		v2 := p.GetVarValue(instrT.Params[v1p])
		v3 := p.GetVarValue(instrT.Params[v1p+1])
		v4 := p.GetVarValue(instrT.Params[v1p+2])

		vs := p.ParamsToStrs(instrT, v1p+3)

		rs := tk.DownloadFile(tk.ToStr(v2), tk.ToStr(v3), tk.ToStr(v4), vs...)

		p.SetVarInt(pr, rs)

		return ""

	case 20291: // getResource
		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		p.SetVarInt(pr, tk.SafelyGetStringForKeyWithDefault(ResourceG, strings.ReplaceAll(tk.ToStr(p.GetVarValue(instrT.Params[v1p])), "~~~", "`"), ""))

		return ""

	case 20293: // getResourceList
		pr := -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			// v1p = 1
		}

		aryT := make([]string, 0, len(ResourceG))
		for k, _ := range ResourceG {
			aryT = append(aryT, k)
		}

		p.SetVarInt(pr, aryT)

		return ""

	case 20310: // htmlToText
		// var v2 []string
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		// if instrT.ParamLen < 3 {
		// 	v2 = []string{}
		// } else {
		// 	v2 = p.GetVarValue(instrT.Params[2]).([]string)
		// }

		v1 := p.GetVarValue(instrT.Params[1])

		v2 := p.ParamsToStrs(instrT, 2)

		rs := tk.HTMLToText(tk.ToStr(v1), v2...)

		p.SetVarInt(pr, rs)

		return ""

	// case 20411: // regReplaceAllStr$
	// 	p1 := p.Pops()
	// 	p2 := p.Pops()
	// 	p3 := p.Pops()

	// 	rs := regexp.MustCompile(p2).ReplaceAllString(p3, p1)

	// 	p.Push(rs)

	// 	return ""

	case 20421: // regFindAll
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		v3 := p.GetVarValue(instrT.Params[v1p+2])

		rs := tk.RegFindAllX(tk.ToStr(v1), tk.ToStr(v2), tk.ToInt(v3, 0))

		p.SetVarInt(pr, rs)

		return ""

	case 20423: // regFind
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		v3 := p.GetVarValue(instrT.Params[v1p+2])

		rs := tk.RegFindFirstX(tk.ToStr(v1), tk.ToStr(v2), tk.ToInt(v3, 0))

		p.SetVarInt(pr, rs)

		return ""

	case 20425: // regFindIndex
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		rs1, rs2 := tk.RegFindFirstIndexX(tk.ToStr(v1), tk.ToStr(v2))

		p.SetVarInt(pr, []int{rs1, rs2})

		return ""

	case 20431: // regMatch
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		rs := tk.RegMatchX(tk.ToStr(v1), tk.ToStr(v2))

		p.SetVarInt(pr, rs)

		return ""

	case 20441: // regContains
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		rs := tk.RegContainsX(tk.ToStr(v1), tk.ToStr(v2))

		p.SetVarInt(pr, rs)

		return ""

	case 20443: // regContainsIn
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2s := p.ParamsToStrs(instrT, v1p+1)

		rs := tk.RegContainsIn(tk.ToStr(v1), v2s...)

		p.SetVarInt(pr, rs)

		return ""

	case 20445: // regCount
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := p.GetVarValue(instrT.Params[v1p+1])

		rs := tk.RegFindAllIndexX(tk.ToStr(v1), tk.ToStr(v2))

		p.SetVarInt(pr, len(rs))

		return ""

	case 20451: // regSplit
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		s1 := tk.ToStr(p.GetVarValue(instrT.Params[1]))

		s2 := tk.ToStr(p.GetVarValue(instrT.Params[2]))

		countT := -1

		if instrT.ParamLen > 3 {
			countT = tk.ToInt(p.GetVarValue(instrT.Params[3]))
		}

		listT := tk.RegSplitX(s1, s2, countT)

		p.SetVarInt(pr, listT)

		return ""

	case 20501: // sleep
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		v1 := tk.ToFloat(p.GetVarValue(instrT.Params[0]))

		tk.Sleep(v1)

		return ""

	case 20511: // getClipText
		var pr int

		if instrT.ParamLen < 1 {
			pr = -5
		} else {
			pr = instrT.Params[0].Ref
		}

		strT := tk.GetClipText()

		p.SetVarInt(pr, strT)

		return ""

	case 20512: // setClipText
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		rsT := tk.SetClipText(tk.ToStr(p.GetVarValue(instrT.Params[v1p])))

		p.SetVarInt(pr, rsT)

		return ""

	case 20521: // getEnv
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		rsT := tk.GetEnv(tk.ToStr(p.GetVarValue(instrT.Params[v1p])))

		p.SetVarInt(pr, rsT)

		return ""

	case 20522: // setEnv
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		rsT := tk.SetEnv(tk.ToStr(p.GetVarValue(instrT.Params[v1p])), tk.ToStr(p.GetVarValue(instrT.Params[v1p+1])))

		p.SetVarInt(pr, rsT)

		return ""

	case 20523: // removeEnv
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		rsT := os.Unsetenv(tk.ToStr(p.GetVarValue(instrT.Params[v1p])))

		p.SetVarInt(pr, rsT)

		return ""

	case 20601: // systemCmd
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		optsA := p.ParamsToStrs(instrT, 2)

		p.SetVarInt(pr, tk.SystemCmd(v1, optsA...))

		return ""

	case 20603: // openWithDefault
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT := tk.RunWinFileWithSystemDefault(v1)

		if rsT != "" {
			rsT = tk.GenerateErrorString(rsT)
		}

		p.SetVarInt(pr, rsT)

		return ""

	case 20901: // getOSName
		pr := -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			// v1p = 1
		}

		p.SetVarInt(pr, runtime.GOOS)

		return ""

	case 21101: // loadText
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int
		v1p := 0

		if instrT.ParamLen < 2 {
			pr = -5
		} else {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		fcT, errT := tk.LoadStringFromFileE(p.GetVarValue(instrT.Params[v1p]).(string))

		if errT != nil {
			p.SetVarInt(pr, errT)
		} else {
			p.SetVarInt(pr, fcT)
		}

		return ""

	case 21103: // saveText
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		rsT := tk.SaveStringToFile(tk.ToStr(p.GetVarValue(instrT.Params[v1p])), tk.ToStr(p.GetVarValue(instrT.Params[v1p+1])))

		if rsT != "" {
			p.SetVarInt(pr, fmt.Errorf(tk.GetErrStr(rsT)))
		} else {
			p.SetVarInt(pr, "")
		}

		return ""

	case 21105: // loadBytes
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		fcT := tk.LoadBytesFromFile(tk.ToStr(p.GetVarValue(instrT.Params[v1p])))

		p.SetVarInt(pr, fcT)

		return ""

	case 21106: // saveBytes
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		v1b, ok := v1.([]byte)

		var rsT interface{}

		if ok {
			rsT = tk.SaveBytesToFileE(v1b, v2)

			p.SetVarInt(pr, rsT)
			return ""
		}

		v1buf, ok := v1.(*bytes.Buffer)

		if ok {
			rsT = tk.SaveBytesToFileE(v1buf.Bytes(), v2)

			p.SetVarInt(pr, rsT)
			return ""
		}

		// p.SetVarInt(pr, fmt.Errorf())

		return p.ErrStrf("无法处理的类型：%T", v1)

	case 21107: // loadBytesLimit
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		fcT := tk.LoadBytesFromFile(tk.ToStr(p.GetVarValue(instrT.Params[1])), tk.ToInt(p.GetVarValue(instrT.Params[2])))

		p.SetVarInt(pr, fcT)

		return ""

	case 21201: // writeStr
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := p.GetVarValue(instrT.Params[v1p])

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		switch nv := v1.(type) {
		case string:
			p.SetVarInt(pr, nv+v2)
			return ""
		case io.StringWriter:
			n, err := nv.WriteString(v2)

			if err != nil {
				p.SetVarInt(pr, err)
				return ""
			}

			p.SetVarInt(pr, n)
			return ""
		default:
			p.SetVarInt(pr, fmt.Errorf("无法处理的类型：%T(%v)", v1, v1))
			return ""

		}

		return ""

	case 21501: // createFile
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		optsA := p.ParamsToStrs(instrT, v1p+1)

		if tk.IfSwitchExistsWhole(optsA, "-return") {
			if !tk.IfSwitchExistsWhole(optsA, "-overwrite") && tk.IfFileExists(v1) {
				p.SetVarInt(pr, fmt.Errorf("文件已存在"))
				return ""
			}

			fileT, errT := os.Create(v1)

			if errT != nil {
				p.SetVarInt(pr, errT)
				return ""
			}

			p.SetVarInt(pr, fileT)
			return ""
		}

		errT := tk.CreateFile(v1, optsA...)

		p.SetVarInt(pr, errT)

		return ""

	case 21503: // openFile
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		vs := p.ParamsToStrs(instrT, v1p+1)

		fileT := tk.OpenFile(v1, vs...)

		p.SetVarInt(pr, fileT)

		return ""

	case 21507: // closeFile
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p]).(*os.File)

		errT := v1.Close()

		p.SetVarInt(pr, errT)

		return ""

	case 21601: // cmpBinFile
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		f1 := tk.ToStr(p.GetVarValue(instrT.Params[1]))
		f2 := tk.ToStr(p.GetVarValue(instrT.Params[2]))

		optsT := p.ParamsToStrs(instrT, 3)

		if tk.IfSwitchExistsWhole(optsT, "-identical") {
			buf1 := tk.LoadBytesFromFile(f1)

			if tk.IsError(buf1) {
				return p.ErrStrf("加载文件（%v）失败：%v", f1, buf1)
			}

			buf2 := tk.LoadBytesFromFile(f2)

			if tk.IsError(buf2) {
				return p.ErrStrf("加载文件（%v）失败：%v", f2, buf2)
			}

			realBuf1 := buf1.([]byte)
			realBuf2 := buf2.([]byte)

			len1 := len(realBuf1)

			len2 := len(realBuf2)

			lenT := len1

			if lenT < len2 {
				lenT = len2
			}

			var c1 int
			var c2 int

			// diffBufT := make([][]int, 0, 100)

			for i := 0; i < lenT; i++ {
				if i >= len1 {
					p.SetVarInt(pr, false)
					return ""
				} else {
					c1 = int(realBuf1[i])
				}

				if i >= len2 {
					p.SetVarInt(pr, false)
					return ""
				} else {
					c2 = int(realBuf2[i])
				}

				if c1 != c2 {
					p.SetVarInt(pr, false)
					return ""
				}

			}

		} else {
			buf1 := tk.LoadBytesFromFile(f1)

			if tk.IsError(buf1) {
				return p.ErrStrf("加载文件（%v）失败：%v", f1, buf1)
			}

			buf2 := tk.LoadBytesFromFile(f2)

			if tk.IsError(buf2) {
				return p.ErrStrf("加载文件（%v）失败：%v", f2, buf2)
			}

			realBuf1 := buf1.([]byte)
			realBuf2 := buf2.([]byte)

			p.SetVarInt(pr, tk.CompareBytes(realBuf1, realBuf2))

			return ""
		}

		p.SetVarInt(pr, true)

		return ""

	case 21701: // fileExists
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT := tk.IfFileExists(v1)

		p.SetVarInt(pr, rsT)

		return ""

	case 21702: // isDir
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT := tk.IsDirectory(v1)

		p.SetVarInt(pr, rsT)

		return ""

	case 21703: // getFileSize
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT, errT := tk.GetFileSize(v1)

		if errT != nil {
			p.SetVarInt(pr, errT)
			return ""
		}

		p.SetVarInt(pr, int(rsT))

		return ""

	case 21705: // getFileInfo
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT, errT := tk.GetFileInfo(v1)

		if errT != nil {
			p.SetVarInt(pr, errT)
			return ""
		}

		p.SetVarInt(pr, rsT)

		return ""

	case 21801: // removeFile
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		if instrT.ParamLen > 2 {
			optsA := p.ParamsToStrs(instrT, 2)

			if tk.IfSwitchExistsWhole(optsA, "-dry") {
				tk.Pl("模拟删除 %v", v1)

				p.SetVarInt(pr, nil)

				return ""
			}
		}

		rsT := tk.RemoveFile(v1)

		p.SetVarInt(pr, rsT)

		return ""

	case 21803: // renameFile
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		if instrT.ParamLen > 3 {
			optsA := p.ParamsToStrs(instrT, 3)

			p.SetVarInt(pr, tk.RenameFile(v1, v2, optsA...))

			return ""

		}

		p.SetVarInt(pr, tk.RenameFile(v1, v2))

		return ""

	case 21805: // copyFile
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		optsA := p.ParamsToStrs(instrT, 3)

		p.SetVarInt(pr, tk.RenameFile(v1, v2, optsA...))

		return ""

	case 21901: // genFileList/getFileList
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr = instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		paramsT := p.ParamsToStrs(instrT, 2)

		rsT := tk.GetFileList(v1, paramsT...)

		p.SetVarInt(pr, rsT)

		return ""

	case 21902: // joinPath
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		rsT := filepath.Join(p.ParamsToStrs(instrT, 1)...)

		p.SetVarInt(pr, rsT)

		return ""

	case 21905: // getCurDir

		pr := -5

		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			// v1p = 1
		}

		rsT := tk.GetCurrentDir()

		p.SetVarInt(pr, rsT)

		return ""

	case 21906: // setCurDir
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		dirT := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, tk.SetCurrentDir(dirT))

		return ""

	case 21907: // getAppDir

		pr := -5

		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			// v1p = 1
		}

		rsT := tk.GetApplicationPath()

		p.SetVarInt(pr, rsT)

		return ""

	case 21910: // extractFileName
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT := filepath.Base(v1)

		p.SetVarInt(pr, rsT)

		return ""

	case 21911: // extractFileExt
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT := filepath.Ext(v1)

		p.SetVarInt(pr, rsT)

		return ""

	case 21912: // extractFileDir
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT := filepath.Dir(v1)

		p.SetVarInt(pr, rsT)

		return ""

	case 21915: // extractPathRel
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		rsT, errT := filepath.Rel(v2, v1)

		if errT != nil {
			p.SetVarInt(pr, errT)
			return ""
		}

		p.SetVarInt(pr, rsT)
		return ""

	case 21921: // ensureMakeDirs
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT := tk.EnsureMakeDirs(v1)

		p.SetVarInt(pr, rsT)

		return ""

	case 22001: // getInput
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT := tk.GetInputf(v1, p.ParamsToList(instrT, v1p+1)...)

		p.SetVarInt(pr, rsT)

		return ""

	case 22003: // getPassword
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rsT := tk.GetInputPasswordf(v1, p.ParamsToList(instrT, v1p+1)...)

		p.SetVarInt(pr, rsT)

		return ""

	case 22101: // toJson
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var pr int = instrT.Params[0].Ref

		argsT := p.ParamsToStrs(instrT, 2)

		vT := tk.ToJSONX(p.GetVarValue(instrT.Params[1]), argsT...)

		p.SetVarInt(pr, vT)

		return ""

	case 22102: // fromJson
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int
		v1p := 0

		if instrT.ParamLen < 2 {
			pr = -5
		} else {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		vT, errT := tk.FromJSON(p.GetVarValue(instrT.Params[v1p]).(string))

		if errT != nil {
			vT = errT
		}

		p.SetVarInt(pr, vT)

		return ""

	case 22201: // toXml
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var p1 int = instrT.Params[0].Ref

		argsT := p.ParamsToList(instrT, 2)

		vT := tk.ToXML(p.GetVarValue(instrT.Params[1]), argsT...)

		p.SetVarInt(p1, vT)

		return ""

	case 23000: // randomize
		if instrT.ParamLen > 0 {
			tk.Randomize(tk.ToInt(p.GetVarValue(instrT.Params[0])))
		} else {
			tk.Randomize()
		}

		return ""

	case 23001: // getRandomInt
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		minT := 0
		maxT := tk.MAX_INT

		pr := instrT.Params[0].Ref

		if instrT.ParamLen > 2 {
			minT = tk.ToInt(p.GetVarValue(instrT.Params[1]))
			maxT = tk.ToInt(p.GetVarValue(instrT.Params[2]))
		} else {
			maxT = tk.ToInt(p.GetVarValue(instrT.Params[1]))
		}

		rs := tk.GetRandomIntInRange(minT, maxT)

		p.SetVarInt(pr, rs)

		return ""

	case 23003: // genRandomFloat
		pr := -5
		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
		}

		p.SetVarInt(pr, rand.Float64())

		return ""

	case 23101: // genRandomStr
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		p1 := instrT.Params[0].Ref

		listT := p.ParamsToStrs(instrT, 1)

		rs := tk.GenerateRandomStringX(listT...)

		p.SetVarInt(p1, rs)

		return ""

	case 24101: // md5
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		p.SetVarInt(pr, tk.MD5Encrypt(tk.ToStr(v1)))

		return ""

	case 24201: // simpleEncode
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		var v2 byte = '_'
		if instrT.ParamLen > 2 {
			v2 = p.GetVarValue(instrT.Params[2]).(byte)
		}

		rsT := tk.EncodeStringCustomEx(v1, v2)

		p.SetVarInt(pr, rsT)

		return ""

	case 24203: // simpleDecode
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		var v2 byte = '_'
		if instrT.ParamLen > 2 {
			v2 = p.GetVarValue(instrT.Params[2]).(byte)
		}

		rsT := tk.DecodeStringCustom(v1, v2)

		p.SetVarInt(pr, rsT)

		return ""

	case 24301: // urlEncode
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		var v2 string = ""
		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(instrT.Params[2]))
		}

		if v2 == "-method=x" {
			p.SetVarInt(pr, tk.UrlEncode2(v1))
		} else {
			p.SetVarInt(pr, tk.UrlEncode(v1))
		}

		return ""

	case 24303: // urlDecode
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rStrT, errT := url.QueryUnescape(v1)
		if errT != nil {
			p.SetVarInt(pr, errT)
		} else {
			p.SetVarInt(pr, rStrT)
		}

		return ""

	case 24401: // base64Encode
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])

		p.SetVarInt(pr, tk.ToBase64(v1))

		return ""

	case 24403: // base64Decode
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rStrT, errT := base64.StdEncoding.DecodeString(v1)
		if errT != nil {
			p.SetVarInt(pr, errT)
		} else {
			p.SetVarInt(pr, rStrT)
		}

		return ""

	case 24501: // htmlEncode
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, tk.EncodeHTML(v1))

		return ""

	case 24503: // htmlDecode
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, tk.DecodeHTML(v1))

		return ""

	case 24601: // hexEncode
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, tk.StrToHex(v1))

		return ""

	case 24603: // hexDecode
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		if strings.HasPrefix(v1, "HEX_") {
			v1 = v1[4:]
		}

		p.SetVarInt(pr, tk.HexToStr(v1))

		return ""

	case 24801: // toUtf8/toUTF8
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := p.GetVarValue(instrT.Params[v1p])
		var v2 string = ""

		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
		}

		var rs interface{}

		switch nv := v1.(type) {
		case string:
			rs = tk.ConvertStringToUTF8(nv, v2)
		case []byte:
			rs = tk.ConvertToUTF8(nv, v2)
		default:
			return p.ErrStrf("参数类型错误：%T", v1)
		}

		p.SetVarInt(pr, rs)

		return ""

	case 25101: // encryptText
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		v2 := ""
		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
		}

		rsT := tk.EncryptStringByTXDEF(v1, v2)

		if tk.IsErrStr(rsT) {
			p.SetVarInt(pr, fmt.Errorf(tk.GetErrStr(rsT)))
		} else {
			p.SetVarInt(pr, rsT)
		}

		return ""

	case 25103: // decryptText
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		v2 := ""
		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
		}

		rsT := tk.DecryptStringByTXDEF(v1, v2)

		if tk.IsErrStr(rsT) {
			p.SetVarInt(pr, fmt.Errorf(tk.GetErrStr(rsT)))
		} else {
			p.SetVarInt(pr, rsT)
		}

		return ""

	case 25201: // encryptData
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p]).([]byte)

		v2 := ""
		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(instrT.Params[2]))
		}

		rsT := tk.EncryptDataByTXDEF(v1, v2)

		p.SetVarInt(pr, rsT)

		return ""

	case 25203: // decryptData
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p]).([]byte)

		v2 := ""
		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(instrT.Params[2]))
		}

		rsT := tk.DecryptDataByTXDEF(v1, v2)

		p.SetVarInt(pr, rsT)

		return ""

	case 26001: // getRandomPort
		pr := -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			// v1p = 1
		}

		listener, errT := net.Listen("tcp", ":0")
		if errT != nil {
			return p.ErrStrf("获取随机端口失败：%v", errT)
		}

		portT := listener.Addr().(*net.TCPAddr).Port
		// fmt.Println("Using port:", portT)
		listener.Close()

		p.SetVarInt(pr, portT)

		return ""

	case 32101: // dbConnect
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		dbT := sqltk.ConnectDBX(tk.ToStr(p.GetVarValue(instrT.Params[v1p])), tk.ToStr(p.GetVarValue(instrT.Params[v1p+1])))

		p.SetVarInt(pr, dbT)

		return ""

	case 32102: // dbClose
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		var pr int = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		errT := sqltk.CloseDBX(p.GetVarValue(instrT.Params[v1p]).(*sql.DB))

		p.SetVarInt(pr, errT)

		return ""

	case 32103: // dbQuery
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		p1 := instrT.Params[0].Ref

		v2 := p.GetVarValue(instrT.Params[1])

		v3 := p.GetVarValue(instrT.Params[2])

		listT := p.ParamsToList(instrT, 3)

		rs := sqltk.QueryDBX(v2.(*sql.DB), tk.ToStr(v3), listT...)

		if tk.IsError(rs) {
			p.SetVarInt(p1, rs)
		} else {
			p.SetVarInt(p1, rs)
		}

		return ""

	case 32105: // dbExec
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		p1 := instrT.Params[0].Ref

		v2 := p.GetVarValue(instrT.Params[1])

		v3 := p.GetVarValue(instrT.Params[2])

		listT := p.ParamsToList(instrT, 3)

		rs := sqltk.ExecDBX(v2.(*sql.DB), tk.ToStr(v3), listT...)

		if tk.IsError(rs) {
			p.SetVarInt(p1, rs)
		} else {
			nv := rs.([]int64)
			p.SetVarInt(p1, []interface{}{int(nv[0]), int(nv[1])})
		}

		return ""

	case 40001: // renderMarkdown
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, tk.RenderMarkdown(v1))

		return ""

	case 41101: // pngEncode
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := p.GetVarValue(instrT.Params[v1p])
		v2 := p.GetVarValue(instrT.Params[v1p+1]).(image.Image)

		v1w, ok := v1.(io.Writer)
		if ok {
			errT := png.Encode(v1w, v2)

			p.SetVarInt(pr, errT)
			return ""

		}

		v1s := tk.ToStr(v1)

		fileT, errT := os.Create(v1s)

		if errT != nil {
			p.SetVarInt(pr, errT)
			return ""
		}

		defer fileT.Close()

		errT = png.Encode(fileT, v2)

		p.SetVarInt(pr, errT)
		return ""

	case 41103: // jpegEnocde/jpgEncode
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := p.GetVarValue(instrT.Params[v1p]).(io.Writer)
		v2 := p.GetVarValue(instrT.Params[v1p+1]).(image.Image)

		vs := p.ParamsToStrs(instrT, v1p+2)

		qualityT := tk.ToInt(tk.GetSwitch(vs, "-quality=", "90"), 90)

		if qualityT < 1 || qualityT > 100 {
			return p.ErrStrf("质量参数错误（1-100）：%v", qualityT)
		}

		v1w, ok := v1.(io.Writer)
		if ok {
			errT := jpeg.Encode(v1w, v2, &jpeg.Options{Quality: qualityT})

			p.SetVarInt(pr, errT)
			return ""

		}

		v1s := tk.ToStr(v1)

		fileT, errT := os.Create(v1s)

		if errT != nil {
			p.SetVarInt(pr, errT)
			return ""
		}

		defer fileT.Close()

		errT = jpeg.Encode(fileT, v2, &jpeg.Options{Quality: qualityT})

		p.SetVarInt(pr, errT)
		return ""

	case 45001: // getActiveDisplayCount
		pr := -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			// v1p = 1
		}

		p.SetVarInt(pr, screenshot.NumActiveDisplays())

		return ""

	case 45011: // getScreenResolution
		pr := instrT.Params[0].Ref

		vs := p.ParamsToStrs(instrT, 1)

		formatT := p.GetSwitchVarValue(vs, "-format=", "")

		idxStrT := p.GetSwitchVarValue(vs, "-index=", "0")

		idxT := tk.StrToInt(idxStrT, 0)

		rectT := screenshot.GetDisplayBounds(idxT)

		var vr interface{}

		if formatT == "" {
			vr = []interface{}{rectT.Max.X, rectT.Max.Y}
		} else if formatT == "raw" || formatT == "rect" {
			vr = rectT
		} else if formatT == "json" {
			vr = tk.ToJSONX(rectT, "-sort")
		}

		// return []interface{}{rectT.Max.X, rectT.Max.Y}
		p.SetVarInt(pr, vr)

		return ""

	case 45021: // captureDisplay
		// pr := -5
		// v1p := 0

		// if instrT.ParamLen > 0 {
		pr := instrT.Params[0].Ref
		v1p := 1
		// }

		v1 := tk.ToInt(p.GetVarValue(GetVarRefFromArray(instrT.Params, v1p)), 0)

		imageA, errT := screenshot.CaptureDisplay(v1)

		if errT != nil {
			p.SetVarInt(pr, errT)
			return ""
		}

		p.SetVarInt(pr, imageA)
		return ""

	case 45023: // captureScreen/captureScreenRect
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}
		// pr := -5
		// v1p := 0

		// if instrT.ParamLen > 0 {
		pr := instrT.Params[0].Ref
		v1p := 1
		// }

		var v1, v2, v3, v4 int

		var errT error

		if instrT.ParamLen > 1 {
			v1 = tk.ToInt(p.GetVarValue(instrT.Params[v1p]), 0)
		} else {
			v1 = 0
		}

		if instrT.ParamLen > 2 {
			v2 = tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]), 0)
		} else {
			v2 = 0
		}

		if instrT.ParamLen > 3 {
			v3 = tk.ToInt(p.GetVarValue(instrT.Params[v1p+2]), 0)
		} else {
			v3 = 0
		}

		if instrT.ParamLen > 4 {
			v4 = tk.ToInt(p.GetVarValue(instrT.Params[v1p+3]), 0)
		} else {
			v4 = 0
		}

		var imageT *image.RGBA

		if v3 == 0 && v4 == 0 {
			imageT, errT = screenshot.CaptureDisplay(0)

		} else {
			imageT, errT = screenshot.Capture(v1, v2, v3, v4)
		}

		if errT != nil {
			p.SetVarInt(pr, errT)
			return ""
		}

		p.SetVarInt(pr, imageT)
		return ""

	case 50001: // genToken
		if instrT.ParamLen < 4 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
		v3 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))

		p.SetVarInt(pr, tk.GenerateToken(v1, v2, v3, p.ParamsToStrs(instrT, 4)...))

		return ""

	case 50003: // checkToken
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, tk.CheckToken(v1, p.ParamsToStrs(instrT, 2)...))

		return ""

	case 60001: // runCode
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		var inputT interface{}

		if instrT.ParamLen > 2 {
			inputT = p.GetVarValue(instrT.Params[v1p+1])
		} else {
			inputT = []interface{}{}
		}

		var rs interface{}

		if instrT.ParamLen > 3 {
			v3 := p.GetVarValue(instrT.Params[v1p+2])

			nv, ok := v3.([]string)

			if ok {
				rs = RunCode(v1, inputT, p.SharedMapM, nv...)
			} else {
				rs = RunCode(v1, inputT, p.SharedMapM, p.ParamsToStrs(instrT, v1p+2)...)
			}
		} else {
			rs = RunCode(v1, inputT, p.SharedMapM)
		}

		p.SetVarInt(pr, rs)

		return ""

	case 70001: // leClear
		leClear()

		return ""

	case 70003: // leLoadStr/leSetAll
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		// pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			// pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		leLoadString(v1)

		// p.SetVarInt(pr, rs)

		return ""

	case 70007: // leSaveStr/leGetAll
		// if instrT.ParamLen < 1 {
		// 	return p.ErrStrf("参数不够")
		// }

		pr := -5
		// v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			// v1p = 1
		}

		// v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// leLoadString(v1)

		p.SetVarInt(pr, leSaveString())

		return ""

	case 70011: // leLoad/leLoadFile
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rs := leLoadFile(v1)

		p.SetVarInt(pr, rs)

		return ""

	case 70012: // leLoadClip
		// if instrT.ParamLen < 1 {
		// 	return p.ErrStrf("参数不够")
		// }

		pr := -5
		// v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			// v1p = 1
		}

		// v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rs := leLoadClip()

		p.SetVarInt(pr, rs)

		return ""

	case 70015: // leLoadSSH
		if instrT.ParamLen < 5 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		pa := p.ParamsToStrs(instrT, v1p)

		var v1, v2, v3, v4, v5 string

		v1 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Host")
		v2 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Port")
		v3 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "User")
		v4 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Password")
		v5 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Path")

		v1 = p.GetSwitchVarValue(pa, "-host=", v1)     // tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 = p.GetSwitchVarValue(pa, "-port=", v2)     // tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
		v3 = p.GetSwitchVarValue(pa, "-user=", v3)     // tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))
		v4 = p.GetSwitchVarValue(pa, "-password=", v4) // tk.ToStr(p.GetVarValue(instrT.Params[v1p+3]))
		if strings.HasPrefix(v4, "740404") {
			v4 = tk.DecryptStringByTXDEF(v4)
		}
		v5 = p.GetSwitchVarValue(pa, "-path=", v5) // tk.ToStr(p.GetVarValue(instrT.Params[v1p+4]))

		sshT, errT := tk.NewSSHClient(v1, v2, v3, v4)

		if errT != nil {
			p.SetVarInt(pr, errT)
			if !leSilentG {
				tk.Pl("连接服务器失败：%v", errT)
			}

			return ""
		}

		defer sshT.Close()

		basePathT, errT := tk.EnsureBasePath("xie")
		if errT != nil {
			p.SetVarInt(pr, errT)
			if !leSilentG {
				tk.Pl("谢语言根路径不存在")
			}
			return ""
		}

		tmpFileT := filepath.Join(basePathT, "leSSHTmp.txt")

		errT = sshT.Download(v5, tmpFileT)

		if errT != nil {
			p.SetVarInt(pr, errT)
			if !leSilentG {
				tk.Pl("从服务器读取文件失败（%v）：%v", v5, errT)
			}
			return ""
		}

		leSSHInfoG["Host"] = v1
		leSSHInfoG["Port"] = v2
		leSSHInfoG["User"] = v3
		leSSHInfoG["Password"] = v4
		leSSHInfoG["Path"] = v5

		errT = leLoadFile(tmpFileT)
		if errT != nil {
			p.SetVarInt(pr, errT)
			if !leSilentG {
				tk.Pl("加载文件失败（%v）：%v", tmpFileT, errT)
			}
			return ""
		}

		p.SetVarInt(pr, errT)

		return ""

	case 70017: // leSave/leSaveFile
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rs := leSaveFile(v1)

		p.SetVarInt(pr, rs)

		return ""

	case 70023: // leSaveClip
		// if instrT.ParamLen < 1 {
		// 	return p.ErrStrf("参数不够")
		// }

		pr := -5
		// v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			// v1p = 1
		}

		rs := leSaveClip()

		p.SetVarInt(pr, rs)

		return ""
	case 70025: // leSaveSSH
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		pa := p.ParamsToStrs(instrT, v1p)

		var v1, v2, v3, v4, v5 string

		v1 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Host")
		v2 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Port")
		v3 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "User")
		v4 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Password")
		v5 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Path")

		v1 = p.GetSwitchVarValue(pa, "-host=", v1)     // tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 = p.GetSwitchVarValue(pa, "-port=", v2)     // tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
		v3 = p.GetSwitchVarValue(pa, "-user=", v3)     // tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))
		v4 = p.GetSwitchVarValue(pa, "-password=", v4) // tk.ToStr(p.GetVarValue(instrT.Params[v1p+3]))
		if strings.HasPrefix(v4, "740404") {
			v4 = tk.DecryptStringByTXDEF(v4)
		}
		v5 = p.GetSwitchVarValue(pa, "-path=", v5) // tk.ToStr(p.GetVarValue(instrT.Params[v1p+4]))

		// tk.Plvsr(v1, v2, v3, v4, v5)

		sshT, errT := tk.NewSSHClient(v1, v2, v3, v4)

		if errT != nil {
			p.SetVarInt(pr, errT)
			if !leSilentG {
				tk.Pl("连接服务器失败：%v", errT)
			}
			return ""
		}

		defer sshT.Close()

		basePathT, errT := tk.EnsureBasePath("xie")
		if errT != nil {
			p.SetVarInt(pr, errT)
			if !leSilentG {
				tk.Pl("谢语言根路径不存在")
			}
			return ""
		}

		tmpFileT := filepath.Join(basePathT, "leSSHTmp.txt")

		errT = leSaveFile(tmpFileT)
		if errT != nil {
			p.SetVarInt(pr, errT)
			if !leSilentG {
				tk.Pl("保存临时文件失败：%v", errT)
			}
			return ""
		}

		errT = sshT.Upload(tmpFileT, v5, pa...)

		if errT != nil {
			p.SetVarInt(pr, errT)
			if !leSilentG {
				tk.Pl("保存文件到服务器失败（%v）：%v", v5, errT)
			}

			return ""
		}

		p.SetVarInt(pr, errT)

		return ""

	case 70016: // leLoadUrl
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rs := leLoadUrl(v1)

		p.SetVarInt(pr, rs)

		return ""

	case 70027: // leInsert
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		rs := leInsertLine(v1, v2)

		p.SetVarInt(pr, rs)

		return ""

	case 70029: // leAppend/leAppendLine
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		rs := leAppendLine(v1)

		p.SetVarInt(pr, rs)

		return ""

	case 70033: // leSet/leSetLine
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		rs := leSetLine(v1, v2)

		p.SetVarInt(pr, rs)

		return ""

	case 70037: // leSetLines
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))
		v3 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))

		rs := leSetLines(v1, v2, v3)

		p.SetVarInt(pr, rs)

		return ""

	case 70039: // leRemove/leRemoveLine
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(instrT.Params[v1p]))

		rs := leRemoveLine(v1)

		p.SetVarInt(pr, rs)

		return ""

	case 70043: // leRemoveLines
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))

		rs := leRemoveLines(v1, v2)

		p.SetVarInt(pr, rs)

		return ""

	case 70045: // leViewAll
		// if instrT.ParamLen < 1 {
		// 	return p.ErrStrf("参数不够")
		// }

		pr := -5
		v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		vs := p.ParamsToStrs(instrT, v1p)

		rs := leViewAll(vs...)

		if tk.IsError(rs) {
			if !leSilentG {
				tk.Pl("内部行编辑器操作失败：%v", rs)
			}
		}

		p.SetVarInt(pr, rs)

		return ""

	case 70047: // leView/leViewLine
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(instrT.Params[v1p]))

		rs := leViewLine(v1)

		if tk.IsError(rs) {
			if !leSilentG {
				tk.Pl("内部行编辑器操作失败：%v", rs)
			}
		}

		p.SetVarInt(pr, rs)

		return ""

	case 70049: // leSort
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToBool(p.GetVarValue(instrT.Params[v1p]))

		rs := leSort(v1)

		p.SetVarInt(pr, rs)

		return ""

	case 70051: // leEnc
		// if instrT.ParamLen < 1 {
		// 	return p.ErrStrf("参数不够")
		// }

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		vs := p.ParamsToStrs(instrT, v1p)

		rs := leConvertToUTF8(vs...)

		p.SetVarInt(pr, rs)

		return ""

	case 70061: // leLineEnd
		// if instrT.ParamLen < 1 {
		// 	return p.ErrStrf("参数不够")
		// }

		pr := instrT.Params[0].Ref
		v1p := 1

		if instrT.ParamLen > 1 {
			v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
			rs := leLineEnd(v1)
			p.SetVarInt(pr, rs)

			return ""
		}

		rs := leLineEnd()

		p.SetVarInt(pr, rs)

		return ""

	case 70071: // leSilent

		pr := instrT.Params[0].Ref
		v1p := 1

		if instrT.ParamLen < 2 {
			p.SetVarInt(pr, leSilent())

			return ""
		}

		rs := leSilent(tk.ToBool(p.GetVarValue(instrT.Params[v1p])))

		p.SetVarInt(pr, rs)

		return ""

	case 70081: // leFind
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		rs := leFind(tk.ToStr(p.GetVarValue(instrT.Params[v1p])))

		if rs != nil {
			if !leSilentG {
				for _, v := range rs {
					tk.Pl("%v", v)
				}
			}
		}

		p.SetVarInt(pr, rs)

		return ""

	case 70083: // leReplace
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		rs := leReplace(v1, v2)

		if rs != nil {
			if !leSilentG {
				for _, v := range rs {
					tk.Pl("%v", v)
				}
				tk.Pl("共替换 %v 处", len(rs))
			}
		}

		p.SetVarInt(pr, rs)

		return ""

	case 70091: // leSSHInfo/leSshInfo
		pr := -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0].Ref
			// v1p = 1
		}

		p.SetVarInt(pr, leSSHInfoG)

		if !leSilentG {
			tk.Pl("leSSHInfo: %v", leSSHInfoG)
		}

		return ""

	case 70098: // leRun
		// if instrT.ParamLen < 2 {
		// 	return p.ErrStrf("参数不够")
		// }

		// pr := -5
		// v1p := 0

		// if instrT.ParamLen > 1 {
		pr := instrT.Params[0].Ref
		v1p := 1
		// }

		v1 := leSaveString()

		var inputT interface{}

		if instrT.ParamLen > 1 {
			inputT = p.GetVarValue(instrT.Params[v1p])
		} else {
			inputT = []interface{}{}
		}

		var rs interface{}

		if instrT.ParamLen > 2 {
			v3 := p.GetVarValue(instrT.Params[v1p+1])

			nv, ok := v3.([]string)

			if ok {
				rs = RunCode(v1, inputT, p.SharedMapM, nv...)
			} else {
				rs = RunCode(v1, inputT, p.SharedMapM, p.ParamsToStrs(instrT, v1p+1)...)
			}
		} else {
			rs = RunCode(v1, inputT, p.SharedMapM)
		}

		p.SetVarInt(pr, rs)

		return ""
	case 80001: // getMimeType
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		extT := filepath.Ext(v1)

		p.SetVarInt(pr, tk.GetMimeTypeByExt(extT))

		return ""

	case 90101: // archiveFilesToZip
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))
		vs := p.ParamsToList(instrT, v1p+1)

		fileNamesT := make([]string, 0, len(vs))
		args1T := make([]string, 0, len(vs))

		for _, vi1 := range vs {
			nvs, ok := vi1.(string)

			if ok {
				if !strings.HasPrefix(nvs, "-") {
					fileNamesT = append(fileNamesT, nvs)
				} else {
					args1T = append(args1T, nvs)
				}

				continue
			}

			nvsa, ok := vi1.([]string)
			if ok {
				for _, vj1 := range nvsa {
					fileNamesT = append(fileNamesT, vj1)
				}

				continue
			}

			nvsi, ok := vi1.([]interface{})
			if ok {
				for _, vj1 := range nvsi {
					fileNamesT = append(fileNamesT, tk.ToStr(vj1))
				}

				continue
			}

		}

		z := &archiver.Zip{
			// CompressionLevel:       3,
			OverwriteExisting: tk.IfSwitchExistsWhole(args1T, "-overwrite"),
			MkdirAll:          tk.IfSwitchExistsWhole(args1T, "-makeDirs"),
			// SelectiveCompression:   true,
			// ImplicitTopLevelFolder: false,
			// ContinueOnError:        false,
		}

		errT := z.Archive(fileNamesT, v1)
		// if errT != nil {
		// 	tk.Plv(errT)
		// 	// tk.AppendStringToFile(tk.Spr("Archive error(%v): %v", filePathA, errT), logFileNameT)
		// 	tk.SaveStringToFile(tk.Spr("Archive error(%v): %v", filePathA, errT), errFileNameT)
		// 	return
		// }

		p.SetVarInt(pr, errT)

		return ""

	case 90111: // extractFilesFromZip
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		destT := "."

		if instrT.ParamLen > 2 {
			destT = tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
		}

		z := &archiver.Zip{
			// CompressionLevel:       3,
			OverwriteExisting: true,
			MkdirAll:          true,
			// SelectiveCompression:   true,
			// ImplicitTopLevelFolder: false,
			// ContinueOnError:        false,
		}

		errT := z.Unarchive(v1, destT)
		// if errT != nil {
		// 	tk.Plv(errT)
		// 	// tk.AppendStringToFile(tk.Spr("Archive error(%v): %v", filePathA, errT), logFileNameT)
		// 	tk.SaveStringToFile(tk.Spr("Archive error(%v): %v", filePathA, errT), errFileNameT)
		// 	return
		// }

		p.SetVarInt(pr, errT)

		return ""

	case 100001: // initWebGUIW/initWebGuiW
		applicationPathT := tk.GetApplicationPath()

		osT := tk.GetOSName()

		if tk.Contains(osT, "inux") {
		} else if tk.Contains(osT, "arwin") {
		} else {
			if !tk.IfFileExists(filepath.Join(applicationPathT, "xiewbr.exe")) {
				// _, errT := exec.LookPath("xiewbr.exe")

				// if errT != nil {
				if !leSilentG {
					tk.Pl("初始化WEB图形界面环境……")
				}

				rs := tk.DownloadFile("http://xie.topget.org/pub/xiewbr.exe", applicationPathT, "xiewbr.exe")

				if tk.IsErrorString(rs) {
					return p.ErrStrf("初始化WEB图形界面环境失败")
				}

				rs = tk.DownloadFile("http://xie.topget.org/pub/scapp.exe", applicationPathT, "scapp.exe")

				if tk.IsErrorString(rs) {
					return p.ErrStrf("初始化WEB图形界面环境失败")
				}
			}
		}

		return ""

	case 100003: // updateWebGuiW
		applicationPathT := tk.GetApplicationPath()

		osT := tk.GetOSName()

		if tk.Contains(osT, "inux") {
		} else if tk.Contains(osT, "arwin") {
		} else {
			if !leSilentG {
				tk.Pl("更新WEB图形界面环境……")
			}

			rs := tk.DownloadFile("http://xie.topget.org/pub/xiewbr.exe", applicationPathT, "xiewbr.exe")

			if tk.IsErrorString(rs) {
				return p.ErrStrf("更新WEB图形界面环境失败")
			}

			rs = tk.DownloadFile("http://xie.topget.org/pub/scapp.exe", applicationPathT, "scapp.exe")

			if tk.IsErrorString(rs) {
				return p.ErrStrf("初始化WEB图形界面环境失败")
			}
		}

		return ""

	case 100011: // initWebGUIC/initWebGuiC
		applicationPathT := tk.GetApplicationPath()

		osT := tk.GetOSName()

		if tk.Contains(osT, "inux") {
		} else if tk.Contains(osT, "arwin") {
		} else {
			zipPathT := filepath.Join(applicationPathT, "xiecbr.zip")
			if !tk.IfFileExists(zipPathT) {
				// _, errT := exec.LookPath("xiewbr.exe")

				// if errT != nil {
				if !leSilentG {
					tk.Pl("初始化WEB图形界面环境（CEF）……")
				}

				rs := tk.DownloadFile("http://xie.topget.org/pub/xiecbr.zip", applicationPathT, "xiecbr.zip")

				if tk.IsErrorString(rs) {
					return p.ErrStrf("初始化WEB图形界面环境失败")
				}

				z := &archiver.Zip{
					// CompressionLevel:       3,
					// OverwriteExisting:      false,
					// MkdirAll:               true,
					// SelectiveCompression:   true,
					// ImplicitTopLevelFolder: false,
					// ContinueOnError:        false,
				}

				errT := z.Unarchive(zipPathT, applicationPathT)
				if errT != nil {
					return p.ErrStrf("解压缩图形环境压缩包失败：%v", errT)
				}

			}
		}

		return ""

	case 200011: // sshUpload
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref
		v1p := 1

		pa := p.ParamsToStrs(instrT, v1p)

		var v1, v2, v3, v4, v5, v6 string

		v1 = strings.TrimSpace(p.GetSwitchVarValue(pa, "-host=", v1))
		v2 = strings.TrimSpace(p.GetSwitchVarValue(pa, "-port=", v2))
		v3 = strings.TrimSpace(p.GetSwitchVarValue(pa, "-user=", v3))
		v4 = strings.TrimSpace(p.GetSwitchVarValue(pa, "-password=", v4))
		if strings.HasPrefix(v4, "740404") {
			v4 = strings.TrimSpace(tk.DecryptStringByTXDEF(v4))
		}
		v5 = strings.TrimSpace(p.GetSwitchVarValue(pa, "-path=", v5))
		v6 = strings.TrimSpace(p.GetSwitchVarValue(pa, "-remotePath=", v6))

		if v1 == "" {
			p.SetVarInt(pr, fmt.Errorf("host不能为空"))
			return ""
		}

		if v2 == "" {
			p.SetVarInt(pr, fmt.Errorf("port不能为空"))
			return ""
		}

		if v3 == "" {
			p.SetVarInt(pr, fmt.Errorf("user不能为空"))
			return ""
		}

		if v4 == "" {
			p.SetVarInt(pr, fmt.Errorf("password不能为空"))
			return ""
		}

		if v5 == "" {
			p.SetVarInt(pr, fmt.Errorf("path不能为空"))
			return ""
		}

		if v6 == "" {
			p.SetVarInt(pr, fmt.Errorf("remotePath不能为空"))
			return ""
		}

		sshT, errT := tk.NewSSHClient(v1, v2, v3, v4)

		if errT != nil {
			p.SetVarInt(pr, errT)

			return ""
		}

		defer sshT.Close()

		errT = sshT.Upload(v5, v6, pa...)

		if errT != nil {
			p.SetVarInt(pr, errT)

			return ""
		}

		p.SetVarInt(pr, nil)
		return ""

		// case 210011: // guiInit
		// 	p.GuiM = make(map[string]interface{}, 10)

		// 	return ""
		// case 210013: // guiNewApp
		// 	pr := -5

		// 	if instrT.ParamLen > 0 {
		// 		pr = instrT.Params[0].Ref
		// 	}

		// 	fontPaths := findfont.List()
		// 	for _, path := range fontPaths {
		// 		// fmt.Println(path)
		// 		//楷体:simkai.ttf
		// 		//黑体:simhei.ttf
		// 		if strings.Contains(path, "simhei.ttf") {
		// 			os.Setenv("FYNE_FONT", path)
		// 			break
		// 		}
		// 	}

		// 	a := app.New()

		// 	p.SetVarInt(pr, a) // Z...
		// 	return ""

		// case 210021: // guiSetFont
		// 	// pr := -5
		// 	v1p := 0

		// 	if instrT.ParamLen > 1 {
		// 		// pr = instrT.Params[0].Ref
		// 		v1p = 1
		// 	}

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	os.Setenv("FYNE_FONT", v1)

		// 	// p.SetVarInt(pr, a)

		// 	return ""

		// case 210031: // guiNewWindow
		// 	pr := -5
		// 	v1p := 0

		// 	if instrT.ParamLen > 4 {
		// 		pr = instrT.Params[0].Ref
		// 		v1p = 1
		// 	}

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))
		// 	v3 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+2]))

		// 	v4 := p.GetVarValue(instrT.Params[v1p+3])

		// 	vc, ok := v4.(int)

		// 	if !ok {
		// 		vc = 0

		// 		vs := tk.ToStr(v4)

		// 		if strings.Contains(vs, "otResiz") {
		// 			vc += int(giu.MasterWindowFlagsNotResizable)
		// 		}

		// 		if strings.Contains(vs, "float") {
		// 			vc += int(giu.MasterWindowFlagsFloating)
		// 		}

		// 		if strings.Contains(vs, "aximiz") {
		// 			vc += int(giu.MasterWindowFlagsMaximized)
		// 		}

		// 		if strings.Contains(vs, "rameless") {
		// 			vc += int(giu.MasterWindowFlagsFrameless)
		// 		}

		// 		if strings.Contains(vs, "ransparent") {
		// 			vc += int(giu.MasterWindowFlagsTransparent)
		// 		}

		// 	}

		// 	wnd := giu.NewMasterWindow(v1, v2, v3, giu.MasterWindowFlags(vc)) //giu.MasterWindowFlagsNotResizable)

		// 	p.SetVarInt(pr, wnd)

		// 	p.GuiM["window"] = wnd

		// 	return ""

		// case 210032: // guiNewLoop
		// 	pr := instrT.Params[0].Ref
		// 	v1p := 1

		// 	// v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	// v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))
		// 	// v3 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+2]))
		// 	// vs := p.ParamsToList(instrT, v1p)

		// 	lenT := len(instrT.Params)

		// 	sl := make([]giu.Widget, 0, lenT)

		// 	for i := v1p; i < lenT; i++ {
		// 		sl = append(sl, p.GetVarValue(instrT.Params[i]).(giu.Widget))
		// 	}

		// 	// tk.Plv(sl)

		// 	f := func() {
		// 		giu.SingleWindow().Layout(sl...)
		// 	}

		// 	p.SetVarInt(pr, f)

		// 	return ""

		// case 210033: // guiCloseWindow
		// 	// pr := instrT.Params[0].Ref
		// 	// v1p := 0

		// 	var wnd *giu.MasterWindow

		// 	if instrT.ParamLen > 0 {
		// 		wnd = p.GetVarValue(instrT.Params[0]).(*giu.MasterWindow)
		// 	} else {
		// 		wnd = p.GuiM["window"].(*giu.MasterWindow)
		// 	}

		// 	wnd.Close()

		// 	return ""

		// // case 210033: // guiSetContent

		// // 	if instrT.ParamLen < 2 {
		// // 		return p.ErrStrf("参数不够")
		// // 	}

		// // 	v1 := p.GetVarValue(instrT.Params[0]).(fyne.Window)
		// // 	v2 := p.GetVarValue(instrT.Params[1]).(fyne.CanvasObject)

		// // 	v1.SetContent(v2)

		// // 	return ""

		// case 210037: // guiRunLoop

		// 	if instrT.ParamLen < 2 {
		// 		return p.ErrStrf("参数不够")
		// 	}

		// 	v1 := p.GetVarValue(instrT.Params[0]).(*giu.MasterWindow)
		// 	// v2 := p.GetVarValue(instrT.Params[1]).(func())

		// 	// v1.Run(v2)

		// 	pointerT := p.CodePointerM
		// 	v2 := p.GetVarValue(instrT.Params[1])

		// 	tmpPointerT, ok := v2.(int)

		// 	if !ok {
		// 		tmps, ok := v2.(string)

		// 		if !ok {
		// 			return p.ErrStrf("参数类型错误")
		// 		}

		// 		if !strings.HasPrefix(tmps, ":") {
		// 			return p.ErrStrf("标号格式错误：%v", tmps)
		// 		}

		// 		tmps = tmps[1:]

		// 		varIndexT, ok := p.VarIndexMapM[tmps]

		// 		if !ok {
		// 			return p.ErrStrf("无效的标号：%v", tmps)
		// 		}

		// 		tmpPointerT, ok = p.LabelsM[varIndexT]

		// 		if !ok {
		// 			return p.ErrStrf("无效的标号序号：%v(%v)", varIndexT, tmps)
		// 		}

		// 		p.InstrListM[lineA].Params[0].Value = tmpPointerT
		// 		// instrT = VarRef{Ref: instrT.Params[0].Ref, Value: tmpPointerT}
		// 		// tk.Plv(instrT.Params[0])
		// 	}

		// 	// tk.Pln(tmpPointerT)

		// 	beginT := tmpPointerT

		// 	f := func() {
		// 		tmpPointerT = beginT
		// 		// tk.Pln("tmpPointerT", tmpPointerT)
		// 		// for ii := 0; ii < (instrT.ParamLen - 1); ii++ {
		// 		// 	p.Push(p.GetVarValue(instrT.Params[ii+1]))
		// 		// }

		// 		for {
		// 			rs := p.RunLine(tmpPointerT)
		// 			if p.VerboseM {
		// 				if rs != "" && rs != "ret" {
		// 					tk.Pln("rs:", rs)
		// 				}
		// 			}

		// 			nv, ok := rs.(int)

		// 			if ok {
		// 				tmpPointerT = nv
		// 				continue
		// 			}

		// 			nsv, ok := rs.(string)

		// 			if ok {
		// 				if tk.IsErrStr(nsv) {
		// 					if p.VerboseM {
		// 						tk.PlSimpleErrorString(p.ErrStrf("failed to get cmd result: %v", tk.GetErrStr(nsv)))
		// 					}

		// 					return
		// 				}

		// 				if nsv == "ret" {
		// 					return
		// 				} else if nsv == "br" {
		// 					break
		// 				}
		// 			}

		// 			tmpPointerT++
		// 		}

		// 		return

		// 	}

		// 	v1.Run(f)

		// 	return pointerT + 1

		// case 210038: // guiLoopRet

		// 	return "ret"

		// case 210041: // guiNewFunc
		// 	pr := instrT.Params[0].Ref
		// 	v1p := 1

		// 	// v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	// v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))
		// 	// v3 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+2]))
		// 	// vs := p.ParamsToList(instrT, v1p)

		// 	f := func() {
		// 		vmT := NewXie(p.SharedMapM)

		// 		vmT.GuiM = p.GuiM
		// 		// vmT.GuiStrVarsM = p.GuiStrVarsM
		// 		// vmT.GuiIntVarsM = p.GuiIntVarsM
		// 		// vmT.GuiFloatVarsM = p.GuiFloatVarsM

		// 		argCountT := instrT.ParamLen - 2

		// 		for i := 0; i < argCountT; i++ {
		// 			vmT.Push(p.GetVarValue(instrT.Params[v1p+i]))
		// 		}

		// 		vLast := tk.ToStr(p.GetVarValue(instrT.Params[instrT.ParamLen-1]))

		// 		lrs := vmT.Load(vLast)

		// 		if tk.IsErrStr(lrs) {
		// 			if p.VerboseM {
		// 				tk.PlSimpleErrorString(lrs)
		// 			}

		// 			return
		// 		}

		// 		rs := vmT.Run()

		// 		// tk.Plv(rs)

		// 		if tk.IsErrStr(rs) {
		// 			if p.VerboseM {
		// 				tk.PlSimpleErrorString(rs)
		// 			}

		// 			return
		// 		}

		// 		return
		// 	}

		// 	p.SetVarInt(pr, f)

		// 	return ""

		// case 210051: // guiLayout
		// 	if instrT.ParamLen < 2 {
		// 		return p.ErrStrf("参数不够")
		// 	}

		// 	pr := instrT.Params[0].Ref
		// 	v1p := 1

		// 	lenT := len(instrT.Params)

		// 	sl := make([]giu.Widget, 0, lenT)

		// 	for i := v1p + 1; i < lenT; i++ {
		// 		sl = append(sl, p.GetVarValue(instrT.Params[i]).(giu.Widget))
		// 	}

		// 	// tk.Plv(sl)

		// 	v0 := p.GetVarValue(instrT.Params[v1p])
		// 	switch nv := v0.(type) {
		// 	case string:
		// 		if nv == "singleWindow" || nv == "window" {
		// 			giu.SingleWindow().Layout(sl...)
		// 		} else if nv == "singleWindowWithMenuBar" {
		// 			giu.SingleWindowWithMenuBar().Layout(sl...)
		// 		}

		// 		p.SetVarInt(pr, nil)
		// 	default:
		// 		return p.ErrStrf("未知可布局对象类型：%T(%v)", v0, v0)
		// 	}

		// 	// tk.Pln("z")

		// 	return ""

		// case 210101: // guiSetVar
		// 	if instrT.ParamLen < 2 {
		// 		return p.ErrStrf("参数不够")
		// 	}

		// 	v1p := 0

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	v2 := p.GetVarValue(instrT.Params[v1p+1])

		// 	p.GuiM[v1] = v2

		// 	// idxT := tk.ToInt(p.GuiM[v1])

		// 	// p.GuiStrVarsM[idxT] = v2

		// 	// tk.Pl("p.GuiStrVarsM[idxT]: %v", p.GuiStrVarsM[idxT])

		// 	return ""

		// case 210102: // guiSetVarByRef
		// 	if instrT.ParamLen < 3 {
		// 		return p.ErrStrf("参数不够")
		// 	}

		// 	v1p := 0

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	v2 := p.GetVarValue(instrT.Params[v1p+1])

		// 	v3 := p.GetVarValue(instrT.Params[v1p+2])

		// 	p.GuiM[v1] = v2

		// 	switch v2 {
		// 	case "int":
		// 		*(p.GuiM[v1].(*int)) = tk.ToInt(v3)
		// 		break
		// 	case "float":
		// 		*(p.GuiM[v1].(*float64)) = tk.ToFloat(v3)
		// 		break
		// 	case "bool":
		// 		*(p.GuiM[v1].(*bool)) = tk.ToBool(v3)
		// 		break
		// 	case "str":
		// 		*(p.GuiM[v1].(*string)) = tk.ToStr(v3)
		// 		break
		// 	default:
		// 		return p.ErrStrf("未知的数据类型")
		// 	}

		// 	return ""

		// case 210103: // guiNewVar
		// 	if instrT.ParamLen < 2 {
		// 		return p.ErrStrf("参数不够")
		// 	}
		// 	// pr := -5
		// 	v1p := 0

		// 	// if instrT.ParamLen > 1 {
		// 	// 	pr = instrT.Params[0].Ref
		// 	// 	v1p = 1
		// 	// }

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// 	switch v2 {
		// 	case "int":
		// 		p.GuiM[v1] = new(int)
		// 		// p.SetVarInt(pr, new(int))
		// 		break
		// 	case "float":
		// 		p.GuiM[v1] = new(float64)
		// 		// p.SetVarInt(pr, new(float64))
		// 		break
		// 	case "bool":
		// 		p.GuiM[v1] = new(bool)
		// 		// p.SetVarInt(pr, new(bool))
		// 		break
		// 	case "str":
		// 		p.GuiM[v1] = new(string)
		// 		// p.SetVarInt(pr, new(string))
		// 		break
		// 	default:
		// 		return p.ErrStrf("未知的数据类型")
		// 	}

		// 	return ""

		// case 210105: // guiGetVar
		// 	if instrT.ParamLen < 1 {
		// 		return p.ErrStrf("参数不够")
		// 	}

		// 	pr := -5
		// 	v1p := 0

		// 	if instrT.ParamLen > 1 {
		// 		pr = instrT.Params[0].Ref
		// 		v1p = 1
		// 	}

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	p.SetVarInt(pr, p.GuiM[v1])

		// 	return ""

		// case 210107: // guiGetVarByRef
		// 	if instrT.ParamLen < 2 {
		// 		return p.ErrStrf("参数不够")
		// 	}

		// 	pr := -5
		// 	v1p := 0

		// 	if instrT.ParamLen > 2 {
		// 		pr = instrT.Params[0].Ref
		// 		v1p = 1
		// 	}

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// 	switch v2 {
		// 	case "int":
		// 		p.SetVarInt(pr, *(p.GuiM[v1].(*int)))
		// 		break
		// 	case "float":
		// 		p.SetVarInt(pr, *(p.GuiM[v1].(*float64)))
		// 		break
		// 	case "bool":
		// 		p.SetVarInt(pr, *(p.GuiM[v1].(*bool)))
		// 		break
		// 	case "str":
		// 		p.SetVarInt(pr, *(p.GuiM[v1].(*string)))
		// 		break
		// 	default:
		// 		return p.ErrStrf("未知的数据类型")
		// 	}

		// 	return ""

		// // case 211001: // guiNewLabel
		// // 	pr := -5
		// // 	v1p := 0

		// // 	if instrT.ParamLen > 2 {
		// // 		pr = instrT.Params[0].Ref
		// // 		v1p = 1
		// // 	}

		// // 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// // 	v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// // 	// p.GuiStrVarsM = append(p.GuiStrVarsM, v2)

		// // 	p.GuiM[v1] = v2 //len(p.GuiStrVarsM) - 1

		// // 	vr := giu.Label(v2)

		// // 	p.SetVarInt(pr, vr)

		// // 	return ""

		// case 211002: // guiStaticLabel
		// 	pr := -5
		// 	v1p := 0

		// 	if instrT.ParamLen > 1 {
		// 		pr = instrT.Params[0].Ref
		// 		v1p = 1
		// 	}

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	vr := giu.Label(v1)

		// 	p.SetVarInt(pr, vr)

		// 	return ""

		// case 211003: // guiLabel
		// 	pr := -5
		// 	v1p := 0

		// 	if instrT.ParamLen > 1 {
		// 		pr = instrT.Params[0].Ref
		// 		v1p = 1
		// 	}

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	// v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// 	// p.GuiStrVarsM = append(p.GuiStrVarsM, v2)

		// 	// p.GuiM[v1] = v2 //len(p.GuiStrVarsM) - 1

		// 	vr := giu.Label(tk.ToStr(p.GuiM[v1]))

		// 	p.SetVarInt(pr, vr)

		// 	return ""

		// // case 211011: // guiNewButton
		// // 	pr := -5
		// // 	v1p := 0

		// // 	if instrT.ParamLen > 3 {
		// // 		pr = instrT.Params[0].Ref
		// // 		v1p = 1
		// // 	}

		// // 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// // 	v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// // 	v3 := p.GetVarValue(instrT.Params[v1p+2]).(func())

		// // 	p.GuiStrVarsM = append(p.GuiStrVarsM, v2)

		// // 	p.GuiM[v1] = len(p.GuiStrVarsM) - 1

		// // 	vr := giu.Button(func() string { return p.GuiStrVarsM[p.GuiM[v1].(int)] }()).OnClick(v3)

		// // 	p.SetVarInt(pr, vr)

		// // 	return ""

		// case 211013: // guiButton
		// 	pr := -5
		// 	v1p := 0

		// 	if instrT.ParamLen > 2 {
		// 		pr = instrT.Params[0].Ref
		// 		v1p = 1
		// 	}

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	// v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// 	v2 := p.GetVarValue(instrT.Params[v1p+1]).(func())

		// 	// p.GuiStrVarsM = append(p.GuiStrVarsM, v2)

		// 	// p.GuiM[v1] = len(p.GuiStrVarsM) - 1

		// 	vr := giu.Button(tk.ToStr(p.GuiM[v1])).OnClick(v2)

		// 	p.SetVarInt(pr, vr)

		// 	return ""

		// case 211014: // guiStaticButton
		// 	pr := -5
		// 	v1p := 0

		// 	if instrT.ParamLen > 2 {
		// 		pr = instrT.Params[0].Ref
		// 		v1p = 1
		// 	}

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	v2 := p.GetVarValue(instrT.Params[v1p+1]).(func())

		// 	vr := giu.Button(v1).OnClick(v2)

		// 	p.SetVarInt(pr, vr)

		// 	return ""

		// case 211015: // guiInput
		// 	if instrT.ParamLen < 2 {
		// 		return p.ErrStrf("参数不够")
		// 	}

		// 	pr := instrT.Params[0].Ref
		// 	v1p := 1

		// 	v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		// 	vs := p.ParamsToStrs(instrT, v1p+1)

		// 	// v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// 	// v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// 	// p.GuiStrVarsM = append(p.GuiStrVarsM, v2)

		// 	// p.GuiM[v1] = len(p.GuiStrVarsM) - 1

		// 	vr := giu.InputText(p.GuiM[v1].(*string)) //.Label(tk.ToStr(p.GuiM[v2]))

		// 	widthT := tk.GetSwitch(vs, "-width=", "")

		// 	if widthT != "" {
		// 		vr = vr.Size(float32(tk.ToFloat(widthT, 0)))
		// 	}

		// 	labelT := tk.GetSwitch(vs, "-label=", "")

		// 	if labelT != "" {
		// 		vr = vr.Label(labelT)
		// 	}

		// 	hintT := tk.GetSwitch(vs, "-hint=", "")

		// 	if hintT != "" {
		// 		vr = vr.Hint(hintT)
		// 	}

		// 	p.SetVarInt(pr, vr)

		// 	return ""

		// case 211101: // guiRow
		// 	pr := instrT.Params[0].Ref
		// 	v1p := 1

		// 	lenT := len(instrT.Params)

		// 	sl := make([]giu.Widget, 0, lenT)

		// 	for i := v1p; i < lenT; i++ {
		// 		sl = append(sl, p.GetVarValue(instrT.Params[i]).(giu.Widget))
		// 	}
		// 	// v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// 	// p.GuiStrVarsM = append(p.GuiStrVarsM, v2)

		// 	// p.GuiM[v1] = v2 //len(p.GuiStrVarsM) - 1

		// 	vr := giu.Row(sl...)

		// 	// tk.Plv(vr)

		// 	p.SetVarInt(pr, vr)

		// 	return ""

		// case 211103: // guiColumn
		// 	pr := instrT.Params[0].Ref
		// 	v1p := 1

		// 	lenT := len(instrT.Params)

		// 	sl := make([]giu.Widget, 0, lenT)

		// 	for i := v1p; i < lenT; i++ {
		// 		sl = append(sl, p.GetVarValue(instrT.Params[i]).(giu.Widget))
		// 	}
		// 	// v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))

		// 	// p.GuiStrVarsM = append(p.GuiStrVarsM, v2)

		// 	// p.GuiM[v1] = v2 //len(p.GuiStrVarsM) - 1

		// 	vr := giu.Column(sl...)

		// 	// tk.Plv(vr)

		// 	p.SetVarInt(pr, vr)

		// 	return ""

		// case 211105: // guiSpacing/guiGap
		// 	pr := -5
		// 	v1p := 0

		// 	if instrT.ParamLen > 2 {
		// 		pr = instrT.Params[0].Ref
		// 		v1p = 1
		// 	}

		// 	v1 := tk.ToFloat(p.GetVarValue(instrT.Params[v1p]))
		// 	v2 := tk.ToFloat(p.GetVarValue(instrT.Params[v1p+1]))

		// 	vr := giu.Dummy(float32(v1), float32(v2))

		// 	p.SetVarInt(pr, vr)

		// 	return ""

		// case 211101: // guiNewVBox
		// 	pr := instrT.Params[0].Ref
		// 	v1p := 1

		// 	vs := p.ParamsToList(instrT, v1p)

		// 	// var aryT []fyne.CanvasObject = make([]fyne.CanvasObject, 0, len(vs))

		// 	tk.Pln(9)

		// 	vr := container.NewVBox()

		// 	tk.Pln(10)

		// 	for _, v := range vs {
		// 		tk.Plv(v)
		// 		vr.Add(v.(fyne.CanvasObject))
		// 	}

		// 	tk.Pln(11)

		// 	p.SetVarInt(pr, vr)

		// 	tk.Pln(12)

		// 	return ""

		// end of switch
	}

	return p.ErrStrf("未知命令")
}

func (p *XieVM) CallFunc(codeA string, argCountA int) string {
	vmT := NewXie(p.SharedMapM)

	// argCountT := p.Pop()

	// if argCountT == Undefined {
	// 	return tk.ErrStrf()
	// }

	for i := 0; i < argCountA; i++ {
		vmT.Push(p.Pop())
	}

	lrs := vmT.Load(codeA)

	if tk.IsErrStr(lrs) {
		return lrs
	}

	rs := vmT.Run()

	// tk.Plv(rs)

	if !tk.IsErrStr(rs) {
		argCountT := tk.ToInt(rs) // vmT.Pop().(int)

		for i := 0; i < argCountT; i++ {
			p.Push(vmT.Pop())
		}
	}

	return ""
}

func (p *XieVM) GoFunc(codeA string, argCountA int) string {
	vmT := NewXie(p.SharedMapM)

	vmT.VerboseM = true

	// argCountT := p.Pop()

	// if argCountT == Undefined {
	// 	return tk.ErrStrf()
	// }

	for i := 0; i < argCountA; i++ {
		vmT.Push(p.Pop())
	}

	lrs := vmT.Load(codeA)

	if tk.IsErrStr(lrs) {
		return lrs
	}

	go vmT.Run()

	return ""
}

// func (p *XieVM) RunLine(lineA int) interface{} {
// 	lineT := p.CodeListM[lineA]

// 	listT := strings.SplitN(lineT, " ", 2)

// 	cmdT := listT[0]

// 	paramsT := ""

// 	if len(listT) > 1 {
// 		paramsT = strings.TrimSpace(listT[1])
// 	}

// 	if cmdT == "pass" {
// 		return ""
// 	} else if cmdT == "global" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			if p1 == "" {
// 				return p.ErrStrf("参数不够")
// 			}
// 		}

// 		nameT := p.GetName(p1)

// 		if p2 == "" {
// 			p.VarsM[nameT] = ""
// 			return ""
// 		}

// 		valueT := p.GetValue(p2)

// 		if valueT == "bool" {
// 			p.VarsM[nameT] = false
// 		} else if valueT == "int" {
// 			p.VarsM[nameT] = int(0)
// 		} else if valueT == "float" {
// 			p.VarsM[nameT] = float64(0.0)
// 		} else if valueT == "string" {
// 			p.VarsM[nameT] = ""
// 		} else if valueT == "list" {
// 			p.VarsM[nameT] = []interface{}{}
// 		} else if valueT == "strList" {
// 			p.VarsM[nameT] = []string{}
// 		} else if valueT == "map" {
// 			p.VarsM[nameT] = map[string]interface{}{}
// 		} else if valueT == "strMap" {
// 			p.VarsM[nameT] = map[string]string{}
// 		}

// 		return ""
// 	} else if cmdT == "$assign" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = p.Pop()

// 		return ""
// 	} else if cmdT == "assignBool" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		valueT := p.GetValue(p2)

// 		p.GetVars()[nameT] = tk.ToBool(valueT)

// 		return ""
// 	} else if cmdT == "assignFloat" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		valueT := p.GetValue(p2)

// 		p.GetVars()[nameT] = tk.ToFloat(valueT)

// 		return ""
// 	} else if cmdT == "assignStr" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		valueT := p.GetValue(p2)

// 		p.GetVars()[nameT] = tk.ToStr(valueT)

// 		return ""
// 	} else if cmdT == "<i" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("参数不够")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		p.Push(tk.ToInt(s1) < tk.ToInt(s2))

// 		return ""
// 	} else if cmdT == "intAdd" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("参数不够")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		p.Push(tk.ToInt(s1) + tk.ToInt(s2))

// 		return ""
// 	} else if cmdT == "intDiv" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("参数不够")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		p.Push(tk.ToInt(s1) / tk.ToInt(s2))

// 		return ""
// 	} else if cmdT == "regReplaceAllStr" {
// 		p1 := p.Pops()
// 		p2 := p.Pops()
// 		p3 := p.Pops()

// 		rs := regexp.MustCompile(p2).ReplaceAllString(p3, p1)

// 		p.Push(rs)

// 		return ""
// 	} else if cmdT == "pl" {
// 		listT, errT := tk.ParseCommandLine(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("failed to parse paramters")
// 		}

// 		list1T := []interface{}{}

// 		formatT := ""

// 		for i, v := range listT {
// 			if i == 0 {
// 				formatT = v
// 				continue
// 			}
// 			list1T = append(list1T, p.GetValue(v))
// 		}

// 		tk.Pl(formatT, list1T...)

// 		return ""
// 	} else if cmdT == "plv" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			tk.Plv(p.Pop())
// 			return ""
// 			// return p.ErrStrf("参数不够")
// 		}

// 		valueT := p.GetValue(p1)

// 		tk.Plv(valueT)

// 		return ""
// 	} else if cmdT == "popInt" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["popG"] = tk.ToInt(p.Pop())
// 			return ""
// 			// return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToInt(p.Pop())

// 		return ""
// 	} else if cmdT == "popFloat" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["popG"] = tk.ToFloat(p.Pop())
// 			return ""
// 			// return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToFloat(p.Pop())

// 		return ""
// 	} else if cmdT == "popStr" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["popG"] = p.Pop()
// 			return ""
// 			// return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToStr(p.Pop())

// 		return ""
// 	} else if cmdT == "$peek" {
// 		// p1, errT := p.Get1Param(paramsT)
// 		// if errT != nil {
// 		// 	p.VarsM["peekG"] = p.Peek()
// 		// 	return ""
// 		// 	// return p.ErrStrf("参数不够")
// 		// }

// 		nameT := paramsT

// 		p.GetVars()[nameT] = p.Peek()

// 		return ""
// 	} else if cmdT == "peekBool" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["peekG"] = tk.ToBool(p.Peek())
// 			return ""
// 			// return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToBool(p.Peek())

// 		return ""
// 	} else if cmdT == "peekInt" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["peekG"] = tk.ToInt(p.Peek())
// 			return ""
// 			// return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToInt(p.Peek())

// 		return ""
// 	} else if cmdT == "peekFloat" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["peekG"] = tk.ToFloat(p.Peek())
// 			return ""
// 			// return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToFloat(p.Peek())

// 		return ""
// 	} else if cmdT == "peekStr" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["peekG"] = p.Peek()
// 			return ""
// 			// return p.ErrStrf("参数不够")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToStr(p.Peek())

// 		return ""
// 	} else if cmdT == "push" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(p.Pop())
// 			return ""
// 			// return p.ErrStrf("参数不够")
// 		}

// 		valueT := p.GetValue(p1)

// 		p.Push(valueT)

// 		return ""
// 	} else if cmdT == "pushBool" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(tk.ToBool(p.Pop()))
// 			return ""
// 		}

// 		valueT := p.GetValue(p1)

// 		p.Push(tk.ToBool(valueT))

// 		return ""
// 	} else if cmdT == "pushInt" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(tk.ToInt(p.Pop()))
// 			return ""
// 		}

// 		valueT := p.GetValue(p1)

// 		p.Push(tk.ToInt(valueT))

// 		return ""
// 	} else if cmdT == "pushFloat" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(tk.ToFloat(p.Pop()))
// 			return ""
// 		}

// 		valueT := p.GetValue(p1)

// 		p.Push(tk.ToFloat(valueT))

// 		return ""
// 	} else if cmdT == "pushStr" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(tk.ToStr(p.Pop()))
// 			return ""
// 		}

// 		valueT := p.GetValue(p1)

// 		p.Push(tk.ToStr(valueT))

// 		return ""
// 	} else if cmdT == "getParam" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("参数不够")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		paramT := tk.GetParameter(s1.([]string), tk.ToInt(s2))

// 		p.Push(paramT)

// 		return ""
// 	} else if cmdT == "now" {
// 		p1, _ := p.Get1Param(paramsT)

// 		// var timeStrT string

// 		// if p2 == "formal" {
// 		// 	timeStrT = tk.GetNowTimeStringFormal()
// 		// } else {
// 		// 	timeStrT = tk.GetNowTimeString()
// 		// }

// 		if p1 == "" {
// 			p.Push(time.Now())
// 		} else {
// 			s1 := p.GetName(p1)

// 			p.GetVars()[s1] = time.Now()
// 		}

// 		return ""
// 	} else if cmdT == "getWeb" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			if p1 == "" {
// 				return p.ErrStrf("参数不够")
// 			}
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		var listT []interface{} = s2.([]interface{})

// 		// listT = tk.FromJSONWithDefault(tk.ToStr(s2), []interface{}{}).([]interface{})

// 		// if listT == nil {
// 		// 	listT = []interface{}{}
// 		// }

// 		rs := tk.DownloadWebPageX(tk.ToStr(s1), listT...)

// 		p.Push(rs)

// 		return ""
// 	} else if cmdT == "getRuntimeInfo" || cmdT == "getDeInfo" {
// 		p.Push(tk.ToJSONX(p, "-indent", "-sort"))

// 		return ""
// 	}

// 	return p.ErrStrf("unknown command")
// }

func (p *XieVM) RunDefer() interface{} {
	for {
		instrT := p.CurrentFuncContextM.DeferStackM.Pop()

		if instrT == nil {
			break
		}

		nv, ok := instrT.(Instr)

		if !ok {
			return fmt.Errorf("无效的指令：%v", instrT)
		}

		if p.VerboseM {
			tk.Pl("延迟执行：%v", nv)
		}

		rs := p.RunLine(0, nv)

		if tk.IsErrX(rs) {
			return tk.ErrStrf("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStrX(rs))
		}
	}

	return nil
}

func (p *XieVM) Run(posA ...int) string {
	p.CodePointerM = 0
	if len(posA) > 0 {
		p.CodePointerM = posA[0]
	}

	for {
		resultT := p.RunLine(p.CodePointerM)

		c1T, ok := resultT.(int)

		if ok {
			p.CodePointerM = c1T
		} else {
			rs, ok := resultT.(string)

			if !ok {
				return p.ErrStrf("返回结果错误: (%T)%v", resultT, resultT)
			}

			if tk.IsErrStr(rs) {
				return tk.ErrStrf("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStr(rs))
				// tk.Pl("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), p.CodeSourceMapM[p.CodePointerM]+1, tk.GetErrStr(rs))
				// break
			}

			if rs == "" {
				p.CodePointerM++

				if p.CodePointerM >= len(p.CodeListM) {
					break
				}
			} else if rs == "exit" {

				break
				// } else if rs == "cont" {
				// 	return p.ErrStrf("无效指令: %v", rs)
				// } else if rs == "brk" {
				// 	return p.ErrStrf("无效指令: %v", rs)
			} else {
				tmpI := tk.StrToInt(rs)

				if tmpI < 0 {
					return p.ErrStrf("无效指令: %v", rs)
				}

				if tmpI >= len(p.CodeListM) {
					return p.ErrStrf("指令序号超出范围: %v(%v)/%v", tmpI, rs, len(p.CodeListM))
				}

				p.CodePointerM = tmpI
			}

		}

	}

	rsi := p.RunDefer()

	if tk.IsErrX(rsi) {
		return tk.ErrStrf("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStrX(rsi))
	}

	// tk.Pl(tk.ToJSONX(p, "-indent", "-sort"))

	outIndexT, ok := p.VarIndexMapM["outG"]
	if !ok {
		return tk.ErrStrf("no result")
	}

	return tk.ToStr((*p.FuncContextM.VarsM)[p.FuncContextM.VarsLocalMapM[outIndexT]])

}

func RunCode(codeA string, objA interface{}, sharedMapA *tk.SyncMap, optsA ...string) interface{} {
	vmT := NewXie(sharedMapA)

	if len(optsA) > 0 {
		vmT.SetVar("argsG", optsA)
		vmT.SetVar("全局参数", optsA)
	}

	objT, ok := objA.(map[string]interface{})

	if ok {
		for k, v := range objT {
			vmT.SetVar(k, v)
		}
	} else {
		if objA != nil {
			vmT.SetVar("inputG", objA)
			vmT.SetVar("全局输入", objA)
		}
	}

	// if sharedMapA == nil {
	// 	vmT.SharedMapM = tk.NewSyncMap()
	// } else {
	// 	vmT.SharedMapM = sharedMapA
	// }

	lrs := vmT.Load(codeA)

	if tk.IsErrStr(lrs) {
		return tk.ErrStrToErr(lrs)
	}

	// var argsT []string = tk.JSONToStringArray(tk.GetSwitch(optsA, "-args=", "[]"))

	// if argsT != nil {
	// 	vmT.VarsM["argsG"] = argsT
	// } else {
	// 	vmT.VarsM["argsG"] = []string{}
	// }

	rs := vmT.Run()

	return rs
}
