package xie

import (
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
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
)

var VersionG string = "0.2.3"

type UndefinedStruct struct {
	int
}

func (o UndefinedStruct) String() string {
	return "未定义"
}

var Undefined UndefinedStruct = UndefinedStruct{0}

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

	"ref": 210, // 获取变量的引用（取地址）
	"取引用": 210,

	"unref": 211, // 对引用进行解引用
	"解引用":   211,

	"assignRef": 212, // 根据引用进行赋值（将引用指向的变量赋值）
	"引用赋值":      212,

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

	// assign related
	"assign": 401, // 赋值
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

	// "inc$": 802,
	// "inc*": 803,

	"dec": 810, // 将某个整数变量的值减1，省略参数的话将操作弹栈值
	"减一":  810,

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

	"eval": 998, // 计算一个表达式

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
	"addItem": 1110, //数组中添加项
	"增项":      1110,

	"addStrItem": 1111,

	"deleteItem": 1112, //数组中删除项
	"删项":         1112,

	"addItems": 1115, // 数组添加另一个数组的值
	"增多项":      1115,

	"getAnyItem": 1120,
	"任意类型取项":     1120,

	"setAnyItem": 1121,
	"任意类型置项":     1121,

	// "setItem": 1121,
	// "置项":      1121,
	// "getItemX": 1123,
	// "取项X":      1123,
	"getItem": 1123, // 从数组中取项，结果参数不可省略，之后第一个参数为数组对象，第2个为要获取第几项（从0开始），第3个参数可省略，为取不到项时的默认值，省略时返回undefined
	"取项":      1123,

	"setItem": 1124, // 修改数组中某一项的值
	"置项":      1124,
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

	"rangeMap": 1340, // 遍历映射
	"遍历映射":     1340,

	// object related 对象相关

	"new": 1401, // 新建一个数据或对象，第一个参数为结果放入的变量（不可省略），第二个为字符串格式的数据类型或对象名，后面是可选的0-n个参数，目前支持byte、int等

	"method": 1403, // 对特定数据类型执行一定的方法，例如：method $result $str1 trimSet "ab"，将对一个字符串类型的变量str1去掉首尾的a和b字符，结果放入变量result中（注意，该结果参数不可省略，即使该方法没有返回数据，此时可以考虑用$drop）
	"mt":     1403,

	"member": 1405, // 获取特定数据类型的某个成员变量的值，例如：member $result $requestG "Method"，将获得http请求对象的Method属性值（GET、POST等），结果放入变量result中（注意，该结果参数不可省略，即使该方法没有返回数据，此时可以考虑用$drop）
	"mb":     1405,

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

	"strReplace": 1540, // 字符串替换，用法示例：strReplace $result $str1 $find $replacement

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

	"formatTime": 1971,

	// math related 数学相关
	"abs": 2100, // 取绝对值

	// command-line related 命令行相关
	"getParam": 10001, // 获取指定序号的命令行参数，结果参数外第一个参数为list或strList类型，第二个为整数，第三个为默认值（字符串类型），例：getParam $result $argsG 2 ""
	"获取参数":     10001,

	"getSwitch": 10002, // 获取命令行参数中指定的开关参数，结果参数外第一个参数为list或strList类型，第二个为类似“-code=”的字符串，第三个为默认值（字符串类型），例：getSwitch $result $argsG "-code=" ""，将获取命令行中-code=abc的“abc”部分。

	"ifSwitchExists": 10003, // 判断命令行参数中是否有指定的开关参数，结果参数外第一个参数为list或strList类型，第二个为类似“-verbose”的字符串，例：ifSwitchExists $result $argsG "-verbose"，根据命令行中是否含有-verbose返回布尔值true或false
	"switchExists":   10003,

	// print related 输出相关
	"pln": 10410, // 相当于其它语言的println函数
	"输出行": 10410,

	"plo":   10411, // 输出一个变量或数值的类型和值
	"输出值类型": 10411,

	"pl": 10420, // 相当于其它语言的printf函数再多输出一个换行符\n
	"输出": 10420,

	"plv": 10430, // 输出一个变量或数值的值的内部表达形式
	"输出值": 10430,

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

	"hex":   10821, // 16进制编码，对于数字高位在后
	"hexb":  10822, // 16进制编码，对于数字高位在前
	"unhex": 10823, // 16进制解码，结果是一个字节列表
	"toHex": 10824, // 任意数值16进制编码

	"toBool":  10831,
	"toByte":  10835,
	"toRune":  10837,
	"toInt":   10851, // 任意数值转整数，可带一个默认值（转换出错时返回该值），不带的话返回-1
	"toFloat": 10855,
	"toStr":   10861,
	"toTime":  10871,
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

	// http request/response related HTTP请求相关
	"writeResp":       20110, // 写一个HTTP请求的响应
	"setRespHeader":   20111, // 设置一个HTTP请求的响应头，如setRespHeader $responseG "Content-Type" "text/json; charset=utf-8"
	"writeRespHeader": 20112, // 写一个HTTP请求的响应头状态，如writeRespHeader $responseG #i200
	"getReqHeader":    20113, // 获取一个HTTP请求的请求头信息
	"genJsonResp":     20114, // 生成一个JSON格式的响应字符，用法：genJsonResp $result $requestG "success" "Test passed!"，结果格式类似{"Status":"fail", "Value": "network timeout"}，其中Status字段表示响应处理结果状态，一般只有success和fail两种，分别表示成功和失败，如果失败，Value字段中为失败原因，如果成功，Value中为空或需要返回的信息
	"genResp":         20114,

	"newMux":          20121, // 新建一个HTTP请求处理路由对象，等同于 new mux
	"setMuxHandler":   20122, // 设置HTTP请求路由处理函数
	"setMuxStaticDir": 20123, // 设置静态WEB服务的目录，用法示例：setMuxStaticDir $muxT "/static/" "./scripts" ，设置处理路由“/static/”后的URL为静态资源服务，第1个参数为newMux指令创建的路由处理器对象变量，第2个参数是路由路径，第3个参数是对应的本地文件路径，例如：访问 http://127.0.0.1:8080/static/basic.xie，而当前目录是c:\tmp，那么实际上将获得c:\scripts\basic.xie

	"startHttpServer":  20151, // 启动http服务器，用法示例：startHttpServer $resultT ":80" $muxT ；可以后面加-go参数表示以线程方式启动，此时应注意主线程不要退出，否则服务器线程也会随之退出，可以用无限循环等方式保持运行
	"startHttpsServer": 20153, // 启动https(SSL)服务器，用法示例：startHttpsServer $resultT ":443" $muxT /root/server.crt /root/server.key -go

	// web related WEB相关
	"getWeb": 20210, // 发送一个HTTP网络请求，并获取响应结果（字符串格式），getWeb指令除了第一个参数必须是返回结果的变量，第二个参数是访问的URL，其他所有参数都是可选的，method可以是GET、POST等；encoding用于指定返回信息的编码形式，例如GB2312、GBK、UTF-8等；headers是一个JSON格式的字符串，表示需要加上的自定义的请求头内容键值对；参数中还可以有一个映射类型的变量或值，表示需要POST到服务器的参数，用法示例：getWeb $resultT "http://127.0.0.1:80/xms/xmsApi" -method=POST -encoding=UTF-8 -timeout=15 -headers=`{"Content-Type": "application/json"}` $mapT

	// html related HTML相关
	"htmlToText": 20310, // 将HTML转换为字符串，用法示例：htmlToText $result $str1 "flat"，第3个参数开始是可选参数，表示HTML转文本时的选项

	// regex related 正则表达式相关
	// "regReplaceAllStr$": 20411,

	"regFindAll":   20421, // 获取正则表达式的所有匹配，用法示例：regFindAll $result $str1 $regex1 $group
	"regFind":      20423, // 获取正则表达式的第一个匹配，用法示例：regFind $result $str1 $regex1 $group
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

	"getOSName": 20901,

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

	"genFileList": 21901, // 生成目录中的文件列表，即获取指定目录下的符合条件的所有文件，例：getFileList $result `d:\tmp` "-recursive" "-pattern=*" "-exclusive=*.txt" "-withDir" "-verbose"，另有 -compact 参数将只给出Abs、Size、IsDir三项，列表项对象内容类似：map[Abs:D:\tmpx\test1.gox Ext:.gox IsDir:false Mode:-rw-rw-rw- Name:test1.gox Path:test1.gox Size:353339 Time:20210928091734]
	"getFileList": 21901,

	"joinPath": 21902, // 合并文件路径，第一个参数是结果参数不可省略，第二个参数开始要合并的路径
	"合并路径":     21902, // 合并文件路径

	"getCurDir": 21905, // 获取当前工作路径
	"setCurDir": 21906, // 设置当前工作路径

	"getAppDir": 21907, // 获取应用路径（谢语言主程序路径）

	"extractFileName": 21910, // 从文件路径中获取文件名部分
	"extractFileExt":  21911, // 从文件路径中获取文件扩展名（后缀）部分
	"extractFileDir":  21912, // 从文件路径中获取文件目录（路径）部分

	"ensureMakeDirs": 21921,

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

	"randomzie": 23000, // 初始化随机种子

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
	"base64Decode": 24403, // Base64解码

	"htmlEncode": 24501, // HTML编码（&nbsp;等）
	"htmlDecode": 24503, // HTML解码

	"hexEncode": 24601, // 十六进制编码，仅针对字符串
	"hexDecode": 24603, // 十六进制解码，仅针对字符串

	// encrypt/decrypt related 加密/解密相关

	"encryptText": 25101, // 用TXDEF方法加密字符串
	"decryptText": 25103, // 用TXDEF方法解密字符串

	"encryptData": 25201, // 用TXDEF方法加密数据（字节列表）
	"decryptData": 25203, // 用TXDEF方法解密数据（字节列表）

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

	// token related 令牌相关
	"genToken":   50001, // 生成令牌，用法：genToken $result $appCode $userID $userRole -secret=abc，其中可选开关secret是加密秘钥，可省略
	"checkToken": 50003, // 检查令牌，用法：checkToken $result XXXXX -secret=abc -expire=2，其中expire是设置的超时秒数（默认为1440），如果成功，返回类似“appCode|userID|userRole|”的字符串；失败返回TXERROR字符串

	// run code related 运行代码相关
	"runCode": 60001, // 运行一段谢语言代码，在新的虚拟机中执行，除结果参数（不可省略）外，第一个参数是字符串类型的代码（必选，后面参数都是可选），第二个参数为任意类型的传入虚拟机的参数（虚拟机内通过inputG全局变量来获取该参数），后面的参数可以是一个字符串数组类型的变量或者多个字符串类型的变量，虚拟机内通过argsG（字符串数组）来对其进行访问。

	// line editor related 内置行文本编辑器有关
	"leClear":       70001, // 清空行文本编辑器缓冲区，例：leClear()
	"leLoadStr":     70003, // 行文本编辑器缓冲区载入指定字符串内容，例：leLoadStr("abc\nbbb\n结束")
	"leSetAll":      70003, // 等同于leLoadString
	"leSaveStr":     70007, // 取出行文本编辑器缓冲区中内容，例：s = leSaveStr()
	"leGetAll":      70007, // 等同于leSaveStr
	"leLoad":        70011, // 从文件中载入文本到行文本编辑器缓冲区中，例：err = leLoad(`c:\test.txt`)
	"leLoadFile":    70011, // 等同于leLoad
	"leSave":        70017, // 将行文本编辑器缓冲区中内容保存到文件中，例：err = leSave(`c:\test.txt`)
	"leSaveFile":    70017, // 等同于leSave
	"leLoadClip":    70021, // 从剪贴板中载入文本到行文本编辑器缓冲区中，例：err = leLoadClip()
	"leSaveClip":    70023, // 将行文本编辑器缓冲区中内容保存到剪贴板中，例：err = leSaveClip()
	"leInsert":      70025, // 行文本编辑器缓冲区中的指定位置前插入指定内容，例：err = leInsert(3， "abc")
	"leInsertLine":  70027, // 行文本编辑器缓冲区中的指定位置前插入指定内容，例：err = leInsertLine(3， "abc")
	"leAppend":      70029, // 行文本编辑器缓冲区中的指定位置后插入指定内容，例：err = leAppend(3， "abc")
	"leAppendLine":  70031, // 行文本编辑器缓冲区中的指定位置后插入指定内容，例：err = leAppendLine(3， "abc")
	"leSet":         70033, // 设定行文本编辑器缓冲区中的指定行为指定内容，例：err = leSet(3， "abc")
	"leSetLine":     70035, // 设定行文本编辑器缓冲区中的指定行为指定内容，例：err = leSetLine(3， "abc")
	"leSetLines":    70037, // 设定行文本编辑器缓冲区中指定范围的多行为指定内容，例：err = leSetLines(3, 5， "abc\nbbb")
	"leRemove":      70039, // 删除行文本编辑器缓冲区中的指定行，例：err = leRemove(3)
	"leRemoveLine":  70041, // 删除行文本编辑器缓冲区中的指定行，例：err = leRemoveLine(3)
	"leRemoveLines": 70043, // 删除行文本编辑器缓冲区中指定范围的多行，例：err = leRemoveLines(1, 3)
	"leViewAll":     70045, // 查看行文本编辑器缓冲区中的所有内容，例：allText = leViewAll()
	"leView":        70047, // 查看行文本编辑器缓冲区中的指定行，例：lineText = leView(18)
	"leSort":        70049, // 将行文本编辑器缓冲区中的行进行排序，唯一参数表示是否降序排序，例：errT = leSort(true)
	"leEnc":         70051, // 将行文本编辑器缓冲区中的文本转换为UTF-8编码，如果不指定原始编码则默认为GB18030编码

	// end of commands/instructions 指令集末尾
}

type VarRef struct {
	Ref   int // -9 - eval, -8 - pop, -7 - peek, -6 - push, -5 - tmp, -4 - pln, -3 - var(string), -2 - drop, -1 - debug, > 0 normal vars
	Value interface{}
}

type Instr struct {
	Code     int
	ParamLen int
	Params   []VarRef
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

	ErrorHandlerM int

	VerboseM bool
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

func NewXie(globalsA ...map[string]interface{}) *XieVM {
	vmT := &XieVM{}

	vmT.InitVM(globalsA...)

	return vmT
}

func (p *XieVM) InitVM(globalsA ...map[string]interface{}) {
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
		} else {
			return VarRef{-3, s1T} // value(string)
		}
	}
}

func isOperator(strA string) (bool, string) {
	return false, ""
}

func evalSingle(exprA []interface{}) (resultR interface{}) {
	// tk.Plvx(exprA[0])
	// tk.Plvx(exprA[2])
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
				resultR = fmt.Errorf("类型不一致：%T -> %T", exprA[0], exprA[1])
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
		} else if opT == "!=" {
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

func (p *XieVM) EvalExpressionNoGroup(strA string, valuesA *map[string]interface{}) interface{} {
	// strT := strA[1 : len(strA)-1]
	// tk.Pl("EvalExpressionNoGroup: %v", strA)
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

		strA = "$pop"

	}

	listT := strings.Split(strA, " ")

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

		if tk.InStrings(v, "+", "-", "*", "/", "%", "!", "&&", "||", "==", "!=", ">", "<", ">=", "<=", "&", "|", "^", ">>", "<<") {
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

			return (*contextT.VarsM)[nv]
		}

		return Undefined
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

			return (*contextT.VarsM)[nv], (*contextT).Layer
		}

		return Undefined, (*contextT).Layer
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

	if idxT < 0 {
		return Undefined
	}

	contextT := p.FuncContextM

	nv, ok := contextT.VarsLocalMapM[idxT]

	if !ok {
		return Undefined
	}

	return (*contextT.VarsM)[nv]

	// return p.VarsM[idxT]

	// vT, ok := p.VarsM[idxT]

	// if !ok {
	// 	return Undefined
	// }

	// return vT
}

func (p *XieVM) ParseLine(commandA string) ([]string, error) {
	var args []string

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
		return []string{}, fmt.Errorf("Unclosed quote in command line: %v", command)
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
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

	listT, errT := p.ParseLine(v)
	if errT != nil {
		instrT := Instr{Code: InstrNameSet["invalidInstr"], ParamLen: 1, Params: []VarRef{VarRef{Ref: -3, Value: "参数解析失败"}}}
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
		p.CodeSourceMapM[pointerT] = iFirstT
		pointerT++
	}

	for i := originCodeLenT; i < len(p.CodeListM); i++ {
		// listT := strings.SplitN(v, " ", 3)
		v := p.CodeListM[i]
		listT, errT := p.ParseLine(v)
		if errT != nil {
			return p.ErrStrf("参数解析失败")
		}

		lenT := len(listT)

		instrNameT := strings.TrimSpace(listT[0])

		codeT, ok := InstrNameSet[instrNameT]

		if !ok {
			instrT := Instr{Code: codeT, ParamLen: 1, Params: []VarRef{VarRef{Ref: -3, Value: v}}} //&([]VarRef{})}
			p.InstrListM = append(p.InstrListM, instrT)

			return tk.ErrStrf("编译错误(行 %v/%v %v): 未知指令", i, p.CodeSourceMapM[i]+1, tk.LimitString(p.SourceM[p.CodeSourceMapM[i]], 50))
		}

		instrT := Instr{Code: codeT, Params: make([]VarRef, 0, lenT-1)} //&([]VarRef{})}

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
	funcContextT := FuncContext{VarsM: &([]interface{}{}), VarsLocalMapM: make(map[int]int, 10), ReturnPointerM: p.CodePointerM + 1, Layer: p.FuncStackPointerM + 1}

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
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
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
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

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
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

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
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

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
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

	p.Push(vA)
}

func (p *XieVM) GetVarInt(keyA int) interface{} {
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

	contextT := *(p.CurrentFuncContextM)

	localIdxT, ok := contextT.VarsLocalMapM[keyA]

	if !ok {
		return Undefined
	}

	return (*contextT.VarsM)[localIdxT]
}

func (p *XieVM) GetVar(keyA string) interface{} {
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

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
		tk.Pln(`谢语言（Xielang）版本` + VersionG + `

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

	return tk.JoinLines(leBufG)
}

func leLoadFile(fileNameA string) error {
	if leBufG == nil {
		leClear()
	}

	strT, errT := tk.LoadStringFromFileE(fileNameA)

	if errT != nil {
		return errT
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

	textT := tk.JoinLines(leBufG)

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

	textT := tk.JoinLines(leBufG)

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
		textT := tk.JoinLines(leBufG)

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

	leBufG = tk.SplitLines(tk.ConvertStringToUTF8(tk.JoinLines(leBufG), encT))

	return nil
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

		if instrT.ParamLen < 2 {
			p.SetVarIntGlobal(instrT.Params[0].Ref, nil)
			// p.Curr.VarsM[instrT.Params[0].Ref] = ""
			return ""
		}

		valueT := instrT.Params[1].Value

		if valueT == "bool" || valueT == "布尔" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, false)
			// p.VarsM[instrT.Params[0].Ref] = false
		} else if valueT == "int" || valueT == "整数" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, int(0))
			// p.VarsM[instrT.Params[0].Ref] = int(0)
		} else if valueT == "byte" || valueT == "字节" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, byte(0))
			// p.VarsM[instrT.Params[0].Ref] = int(0)
		} else if valueT == "rune" || valueT == "如痕" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, rune(0))
			// p.VarsM[instrT.Params[0].Ref] = int(0)
		} else if valueT == "float" || valueT == "小数" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, float64(0.0))
			// p.VarsM[instrT.Params[0].Ref] = float64(0.0)
		} else if valueT == "str" || valueT == "字符串" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, "")
			// p.VarsM[instrT.Params[0].Ref] = ""
		} else if valueT == "list" || valueT == "列表" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, []interface{}{})
			// p.VarsM[instrT.Params[0].Ref] = []interface{}{}
		} else if valueT == "strList" || valueT == "字符串列表" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, []string{})
			// p.VarsM[instrT.Params[0].Ref] = []string{}
		} else if valueT == "byteList" || valueT == "字节列表" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, []byte{})
		} else if valueT == "runeList" || valueT == "如痕列表" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, []byte{})
		} else if valueT == "map" || valueT == "映射" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, map[string]interface{}{})
			// p.VarsM[instrT.Params[0].Ref] = map[string]interface{}{}
		} else if valueT == "strMap" || valueT == "字符串映射" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, map[string]string{})
			// p.VarsM[instrT.Params[0].Ref] = map[string]string{}
		} else if valueT == "time" || valueT == "时间" {
			p.SetVarIntGlobal(instrT.Params[0].Ref, time.Now())
		} else {
			p.SetVarIntGlobal(instrT.Params[0].Ref, nil)
		}

		return ""
	case 203: // var
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

		nameT := instrT.Params[0].Ref

		// contextT := p.CurrentFuncContextM

		if instrT.ParamLen < 2 {
			p.SetVarIntLocal(nameT, nil)
			// contextT.VarsM[nameT] = ""
			return ""
		}

		valueT := instrT.Params[1].Value

		if valueT == "bool" || valueT == "布尔" {
			p.SetVarIntLocal(nameT, false)
			// varsT[nameT] = false
		} else if valueT == "int" || valueT == "整数" {
			p.SetVarIntLocal(nameT, int(0))
			// varsT[nameT] = int(0)
		} else if valueT == "byte" || valueT == "字节" {
			p.SetVarIntLocal(nameT, byte(0))
		} else if valueT == "rune" || valueT == "如痕" {
			p.SetVarIntLocal(nameT, rune(0))
		} else if valueT == "float" || valueT == "小数" {
			p.SetVarIntLocal(nameT, float64(0.0))
			// varsT[nameT] = float64(0.0)
		} else if valueT == "str" || valueT == "字符串" {
			p.SetVarIntLocal(nameT, "")
			// varsT[nameT] = ""
		} else if valueT == "list" || valueT == "列表" {
			p.SetVarIntLocal(nameT, []interface{}{})
			// varsT[nameT] = []interface{}{}
		} else if valueT == "strList" || valueT == "字符串列表" {
			p.SetVarIntLocal(nameT, []string{})
		} else if valueT == "byteList" || valueT == "字节列表" {
			p.SetVarIntLocal(instrT.Params[0].Ref, []byte{})
		} else if valueT == "runeList" || valueT == "如痕列表" {
			p.SetVarIntLocal(instrT.Params[0].Ref, []byte{})
		} else if valueT == "map" || valueT == "映射" {
			p.SetVarIntLocal(nameT, map[string]interface{}{})
			// varsT[nameT] = map[string]interface{}{}
		} else if valueT == "strMap" || valueT == "字符串映射" {
			p.SetVarIntLocal(nameT, map[string]string{})
			// varsT[nameT] = map[string]string{}
		} else if valueT == "time" || valueT == "时间" {
			p.SetVarIntLocal(nameT, time.Now())
		} else {
			p.SetVarIntLocal(instrT.Params[0].Ref, nil)
		}

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
	case 211: // unref
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

		switch nv := v2.(type) {
		case *interface{}:
			p.SetVarInt(pr, *nv)
		case *byte:
			p.SetVarInt(pr, *nv)
		case *int:
			p.SetVarInt(pr, *nv)
		case *rune:
			p.SetVarInt(pr, *nv)
		case *bool:
			p.SetVarInt(pr, *nv)
		case *string:
			p.SetVarInt(pr, *nv)
		default:
			return p.ErrStrf("无法处理的类型：%T", v2)
		}

		return ""
	case 212: // assignRef
		if instrT.ParamLen < 2 {
			return p.ErrStrf("参数不够")
		}

		// nameT := instrT.Params[0].Ref

		p1 := p.GetVarValue(instrT.Params[0]).(*interface{})

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
			} else if valueTypeT == "list" || valueTypeT == "列表" {
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
			} else if valueTypeT == "list" || valueTypeT == "列表" {
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
			return p.ErrStrf("数据类型不匹配")
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
		default:
			return p.ErrStrf("数据类型不匹配")
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

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p]).(bool)

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

	case 1110: // addItem
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

	case 1123: // getItem
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		var pr = instrT.Params[0].Ref

		v1 := p.GetVarValue(instrT.Params[1])

		v2 := tk.ToInt(p.GetVarValue(instrT.Params[2]))

		switch nv := v1.(type) {
		case []interface{}:
			p.SetVarInt(pr, nv[v2])
		case []bool:
			p.SetVarInt(pr, nv[v2])
		case []int:
			p.SetVarInt(pr, nv[v2])
		case []byte:
			p.SetVarInt(pr, nv[v2])
		case []rune:
			p.SetVarInt(pr, nv[v2])
		case []int64:
			p.SetVarInt(pr, nv[v2])
		case []float64:
			p.SetVarInt(pr, nv[v2])
		case []string:
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

		v2 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+1]))
		v3 := tk.ToInt(p.GetVarValue(instrT.Params[v1p+2]))

		// varsT := p.GetVars()
		switch nv := v1.(type) {
		case []interface{}:
			if v2 < 0 || v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []bool:
			if v2 < 0 || v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []int:
			if v2 < 0 || v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []byte:
			if v2 < 0 || v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []rune:
			if v2 < 0 || v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []int64:
			if v2 < 0 || v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []float64:
			if v2 < 0 || v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVarInt(pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		case []string:
			if v2 < 0 || v2 >= len(nv) {
				return p.ErrStrf("序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 >= len(nv) {
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
		case "mux":
			p.SetVarInt(pr, http.NewServeMux())
		case "int":
			p.SetVarInt(pr, new(int))
		case "byte":
			p.SetVarInt(pr, new(byte))
		case "rune":
			p.SetVarInt(pr, new(rune))
		case "string", "str":
			p.SetVarInt(pr, new(string))
		case "stringBuffer", "strBuf":
			p.SetVarInt(pr, new(strings.Builder))
		case "bool":
			p.SetVarInt(pr, new(bool))
		case "time":
			timeT := time.Now()
			p.SetVarInt(pr, &timeT)
		case "mutex", "lock":
			p.SetVarInt(pr, new(sync.RWMutex))
		case "ssh":
			v2 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+1]))
			v3 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+2]))
			v4 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+3]))
			v5 := tk.ToStr(p.GetVarValue(instrT.Params[v1p+4]))

			sshT, errT := tk.NewSSHClient(v2, v3, v4, v5)

			if errT != nil {
				p.SetVarInt(pr, errT)

				return ""
			}

			p.SetVarInt(pr, sshT)
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

				p.SetVarInt(pr, map[string]interface{}{"Time": nv, "Formal": nv.Format(tk.TimeFormat), "Compact": nv.Format(tk.TimeFormat), "Full": fmt.Sprintf("%v", nv), "Year": nv.Year(), "Month": nv.Month(), "Day": nv.Day(), "Hour": nv.Hour(), "Minute": nv.Minute(), "Second": nv.Second(), "Zone": zoneT, "Offset": offsetT})
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
				p.SetVarInt(pr, fmt.Sprintf("未知方法: %v", v2))
				return p.ErrStrf("未知方法: %v", v2)
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

		p.SetVarInt(pr, fmt.Errorf("未知方法：（%v）%v", v1, v2))

		return ""
	// return p.ErrStrf("未知方法：（%v）%v", v1, v2)
	case 1405: // member/mb
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		v1 := p.GetVarValue(instrT.Params[1])

		v2 := tk.ToStr(p.GetVarValue(instrT.Params[2]))

		switch nv := v1.(type) {
		case *http.Request:
			switch v2 {
			case "Method":
				p.SetVarInt(pr, nv.Method)
				return ""
			case "Proto":
				p.SetVarInt(pr, nv.Proto)
				return ""
			case "Host":
				p.SetVarInt(pr, nv.Host)
				return ""
			case "RemoteAddr":
				p.SetVarInt(pr, nv.RemoteAddr)
				return ""
			case "RequestURI":
				p.SetVarInt(pr, nv.RequestURI)
				return ""
			case "TLS":
				p.SetVarInt(pr, nv.TLS)
				return ""
			case "URL":
				p.SetVarInt(pr, nv.URL)
				return ""
			case "Scheme":
				p.SetVarInt(pr, nv.URL.Scheme)
				return ""
			}

			p.SetVarInt(pr, fmt.Sprintf("未知成员: %v", v2))
			return p.ErrStrf("未知成员: %v", v2)

		case *url.URL:
			switch v2 {
			case "Scheme":
				p.SetVarInt(pr, nv.Scheme)
				return ""
			}

			p.SetVarInt(pr, fmt.Sprintf("未知成员: %v", v2))
			return p.ErrStrf("未知成员: %v", v2)
		}

		p.SetVarInt(pr, fmt.Errorf("未知成员：（%v）%v", v1, v2))
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

		if instrT.ParamLen > 2 {
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

		if instrT.ParamLen > 2 {
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

		v1 := tk.ToStr(p.GetVarValue(instrT.Params[v1p]))

		p.SetVarInt(pr, strings.TrimSpace(v1))

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
		if instrT.ParamLen < 2 {
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

	case 10410: // pln
		list1T := []interface{}{}

		for _, v := range instrT.Params {
			list1T = append(list1T, p.GetVarValue(v))
		}

		fmt.Println(list1T...)

		return ""
	case 10411: // plo
		if instrT.ParamLen < 1 {
			vT := p.Pop()
			tk.Pl("(%T)%v", vT, vT)
			return ""
		}

		valueT := p.GetVarValue(instrT.Params[0])

		tk.Pl("(%T)%v", valueT, valueT)

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
	case 10430: // plv
		if instrT.ParamLen < 1 {
			tk.Plv(p.Pop())
			return ""
			// return p.ErrStrf("参数不够")
		}

		s1 := p.GetVarValue(instrT.Params[0])

		tk.Plv(s1)

		return ""

	case 10440: // plErr
		if instrT.ParamLen < 1 {
			tk.PlErr(p.Pop().(error))
			return ""
			// return p.ErrStrf("参数不够")
		}

		s1 := p.GetVarValue(instrT.Params[0]).(error)

		tk.PlErr(s1)

		return ""

	case 10450: // plErrStr
		if instrT.ParamLen < 1 {
			tk.PlErrString(p.Pop().(string))
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
			v2 := tk.ToTime(p.GetVarValue(instrT.Params[v1p+1]))

			p.SetVarInt(pr, tk.ToTime(v1, v2))

			return ""
		}

		p.SetVarInt(pr, tk.ToTime(v1))

		return ""

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

		pr := -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0].Ref
			v1p = 1
		}

		v1 := p.GetVarValue(instrT.Params[v1p])
		p2 := instrT.Params[v1p+1].Ref

		if tk.IsErrX(v1) {
			if instrT.ParamLen > 2 {
				p.SetVarInt(p2, tk.GetErrStrX(v1))
			}

			p.SetVarInt(pr, true)

			return ""
		}

		if instrT.ParamLen > 2 {
			p.SetVarInt(p2, "")
		}

		p.SetVarInt(pr, false)

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
			}

			var paraMapT map[string]string

			paraMapT = tk.FormToMap(req.Form)

			toWriteT := ""

			vmT := NewXie()

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
			// tk.Pl("old: %v", path.Clean(old))
			// tk.Pl("v2: %v", v2)
			// tk.Pl("trim: %v", strings.TrimPrefix(path.Clean(old), v2))

			name := filepath.Join(v3, strings.TrimPrefix(path.Clean(old), v2))

			// tk.Pl("name: %v", name)

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

		if instrT.ParamLen > 1 {
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

		rsT := tk.SaveBytesToFileE(p.GetVarValue(instrT.Params[v1p]).([]byte), tk.ToStr(p.GetVarValue(instrT.Params[v1p+1])))

		p.SetVarInt(pr, rsT)

		return ""

	case 21107: // loadBytesLimit
		if instrT.ParamLen < 3 {
			return p.ErrStrf("参数不够")
		}

		pr := instrT.Params[0].Ref

		fcT := tk.LoadBytesFromFile(tk.ToStr(p.GetVarValue(instrT.Params[1])), tk.ToInt(p.GetVarValue(instrT.Params[2])))

		p.SetVarInt(pr, fcT)

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

	case 21921: // ensureMakeDirs
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

		rsT := tk.EnsureMakeDirs(v1)

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
			v2 = tk.ToStr(p.GetVarValue(instrT.Params[2]))
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
			v2 = tk.ToStr(p.GetVarValue(instrT.Params[2]))
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
				rs = RunCode(v1, inputT, nv...)
			} else {
				rs = RunCode(v1, inputT, p.ParamsToStrs(instrT, v1p+2)...)
			}
		} else {
			rs = RunCode(v1, inputT)
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
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

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

	case 70011: // leLoad
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

	case 70021: // leLoadClip
		if instrT.ParamLen < 1 {
			return p.ErrStrf("参数不够")
		}

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

		// end of switch
	}

	return p.ErrStrf("未知命令")
}

func (p *XieVM) CallFunc(codeA string, argCountA int) string {
	vmT := NewXie()

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
	vmT := NewXie()

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
// 	} else if cmdT == "getNowStr" {
// 		p1, p2, _ := p.Get2Params(paramsT)

// 		var timeStrT string

// 		if p2 == "formal" {
// 			timeStrT = tk.GetNowTimeStringFormal()
// 		} else {
// 			timeStrT = tk.GetNowTimeString()
// 		}

// 		if p1 == "" {
// 			p.Push(timeStrT)
// 		} else {
// 			s1 := p.GetName(p1)

// 			p.GetVars()[s1] = timeStrT
// 		}

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

	// tk.Pl(tk.ToJSONX(p, "-indent", "-sort"))

	outIndexT, ok := p.VarIndexMapM["outG"]
	if !ok {
		return tk.ErrStrf("no result")
	}

	return tk.ToStr((*p.FuncContextM.VarsM)[p.FuncContextM.VarsLocalMapM[outIndexT]])

}

func RunCode(codeA string, objA interface{}, optsA ...string) interface{} {
	vmT := NewXie()

	if len(optsA) > 0 {
		vmT.SetVar("argsG", optsA)
		vmT.SetVar("全局参数", optsA)
	}

	if objA != nil {
		vmT.SetVar("inputG", objA)
		vmT.SetVar("全局输入", objA)
	}

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
