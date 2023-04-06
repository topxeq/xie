package xie

import (
	"bufio"
	"bytes"
	"compress/flate"
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
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/domodwyer/mailyak"
	"github.com/kbinani/screenshot"
	"github.com/mholt/archiver/v3"
	"github.com/topxeq/awsapi"
	"github.com/topxeq/goph"
	"github.com/topxeq/regexpx"
	"github.com/topxeq/sqltk"
	"github.com/topxeq/tk"

	excelize "github.com/xuri/excelize/v2"

	_ "github.com/denisenkom/go-mssqldb"

	_ "github.com/godror/godror"
	_ "github.com/sijms/go-ora/v2"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var VersionG string = "1.1.9"

func Test() {
	tk.Pl("test")
}

// var InstrCodeSet map[int]string = map[int]string{}

var InstrNameSet map[string]int = map[string]int{

	// internal & debug related
	"invalidInstr": 12, // invalid instruction, use internally to indicate invalid instr(s) while parsing commands

	"version": 100, // get current Xielang version, return a string type value, if the result parameter not designated, it will be put to the global variable $tmp(and it's same for other instructions has result and not variable parameters)

	"pass": 101, // do nothing, useful for placeholder

	"debug": 102,

	"debugInfo": 103, // get the debug info
	"varInfo":   104, // get the information of the variables

	"help": 105, // not implemented

	"onError": 106, // set error handler

	"dumpf": 107,

	"defer": 109, // delay running an instruction, the instruction will be running by order(first in last out) when the function returns or the program exits, or error occurrs

	"deferStack": 110, // get defer stack info

	"isUndef": 111, // 判断变量是否未被声明（定义），第一个结果参数可省略，第二个参数是要判断的变量
	"isDef":   112, // 判断变量是否已被声明（定义），第一个结果参数可省略，第二个参数是要判断的变量
	"isNil":   113, // 判断变量是否是nil，第一个结果参数可省略，第二个参数是要判断的变量

	"test":             121, // for test purpose, check if 2 values are equal
	"testByStartsWith": 122, // for test purpose, check if first string starts with the 2nd
	"testByReg":        123, // for test purpose, check if first string matches the regex pattern defined by the 2nd string

	"typeOf": 131, // 获取变量或数值类型（字符串格式），省略所有参数表示获取看栈值（不弹栈）的类型

	"layer": 141, // 获取变量所处的层级（主函数层级为0，调用的第一个函数层级为1，再嵌套调用的为2，……）

	// -- run code related 运行代码相关

	"loadCode": 151, // 载入字符串格式的谢语言代码到当前虚拟机中（加在最后），出错则返回error对象说明原因

	"loadGel": 152, // 从网络载入谢语言函数（称为gel，凝胶，取其封装的意思），生成compiled对象，一般作为封装函数调用，建议用runCall或goRunCall调用（函数通过准全局变量inputL和outL进行出入参数的交互），出错则返回error对象说明原因；用法：loadGet http://example.com/gel/get1.xie -key=abc123，-key参数可以输入解密密钥，-file参数表示从本地文件读取（默认从远程读取也可以用file://协议从本地读取）

	"compile": 153, // compile a piece of code

	"quickRun": 155, // quick run a piece of code, in a new running context(but same VM, so the global values are accessible), use exit to exit the running context, no return value needed(only erturn error object or "undefined")

	"runCode": 156, // 运行一段谢语言代码，在新的虚拟机中执行，除结果参数（不可省略）外，第一个参数是字符串类型的代码或编译后代码（必选，后面参数都是可选），第二个参数为任意类型的传入虚拟机的参数（虚拟机内通过inputG全局变量来获取该参数），后面的参数可以是一个字符串数组类型的变量或者多个字符串类型的变量（也可以是一个字符串表示命令行），虚拟机内通过argsG（字符串数组）来对其进行访问。返回值是虚拟机正常运行返回值，即$outG或exit加参数的返回值。

	"runPiece": 157, // run a piece of code, in current running context，运行一段谢语言代码，在当前的虚拟机和运行上下文中执行，结果参数可省略，第一个参数是字符串类型的代码或编译后代码。不需要返回值，仅当发生运行错误时返回error对象，否则返回undefined，

	"extractRun":      158, // extract a piece of instrs in a running-context to a new running-context
	"extractCompiled": 159, // extract a piece of instrs in a running-context to a compiled object

	"len": 161, // 获取字符串、列表、映射等的长度，参数全省略表示取弹栈值

	"fatalf": 170, // printf then exit the program(类似pl输出信息后退出程序运行)

	"goto": 180, // jump to the instruction line (often indicated by labels)
	"jmp":  180,

	"wait": 191, // 等待可等待的对象，例如waitGroup或chan，如果没有指定，则无限循环等待（中间会周期性休眠），用于等待用户按键退出或需要静止等待等场景；如果给出一个字符串，则输出字符串后等待输入（回车确认）后继续；如果是整数或浮点数则休眠相应的秒数后继续；

	"exitL":  197, // terminate the program(maybe quickDelegate), can with a return value(same as assign the semi-global value $outL)
	"exitfL": 198, // terminate the program(maybe quickDelegate), can with a return string value assembled like fatalf/sprintf(same as assign the semi-global value $outL), usage: exitfL "error is %v" err1
	"exit":   199, // terminate the program, can with a return value(same as assign the global value $outG)

	// var related
	"global": 201, // define a global variable

	"var": 203, // define a local variable

	"const": 205, // 获取预定义常量

	"nil": 207, // make a variable nil

	"ref": 210, //-> 获取变量的引用（取地址）

	// "refNative": 211,

	"unref": 215, // 对引用进行解引用

	"assignRef": 218, // 根据引用进行赋值（将引用指向的变量赋值）

	// push/peek/pop stack related

	"push": 220, // push any value to stack

	"peek": 222, // peek the value on the top of the stack

	"pop": 224, // pop the value on the top of the stack

	"getStackSize": 230,

	"clearStack": 240,

	"pushRun": 250, // push any value to running context stack

	"peekRun": 252, // peek the value on the top of the running stack

	"popRun": 254, // pop the value on the top of the running stack

	"getRunStackSize": 256,

	"clearRunStack": 258,

	// shared sync map(cross-VM) related 全局同步映射相关（跨虚拟机）
	"getSharedMap":           300, // 获取所有的列表项
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
	"assign": 401, // assignment, from local variable to global, assign value to local if not found
	"=":      401,

	"assignGlobal": 491, // 声明（如果未声明的话）并赋值一个全局变量

	"assignFromGlobal": 492, // 声明（如果未声明的话）并赋值一个局部变量，从全局变量获取值（如果需要的话）

	"assignLocal": 493, // 声明（如果未声明的话）并赋值一个局部变量

	// if/else, switch related
	"if":    610, // usage: if $boolValue1 :labelForTrue :labelForElse
	"ifNot": 611, // usage: if @`$a1 == #i3` :+1 :+2

	"ifEval": 631, // 判断第一个参数（字符串类型）表示的表达式计算结果如果是true，则跳转到指定标号处

	"ifEmpty": 641, // 判断是否是空（值为undefined、nil、false、空字符串、小于等于0的整数或浮点数均会满足条件），是则跳转

	"ifEqual":    643, // 判断是否相等，是则跳转
	"ifNotEqual": 644, // 判断是否不等，是则跳转

	"ifErr":  651, // if error or TXERROR string then ... else ...
	"ifErrX": 651,

	"switch": 691, // 用法：switch $variableOrValue $value1 :label1 $value2 :label2 ... :defaultLabel

	"switchCond": 693, // 用法：switch $condition1 :label1 $condition2 :label2 ... :defaultLabel

	// compare related
	"==": 701, // 判断两个数值是否相等，无参数时，比较两个弹栈值，结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

	"!=": 702, // 判断两个数值是否不等，无参数时，比较两个弹栈值，结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

	"<":  703, // 判断两个数值是否是第一个数值小于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值
	">":  704, // 判断两个数值是否是第一个数值大于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值
	"<=": 705, // 判断两个数值是否是第一个数值小于等于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值
	">=": 706, // 判断两个数值是否是第一个数值大于等于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

	"cmp": 790, // 比较两个数值，根据结果返回-1，0或1，分别表示小于、等于、大于，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

	// operator related

	"inc": 801, // ++
	"++":  801,

	"dec": 810, // --
	"--":  810,

	"add": 901, // add 2 values
	"+":   901,

	"sub": 902, // 两个数值相减，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值
	"-":   902,

	"mul": 903, // 两个数值相乘，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值
	"*":   903,

	"div": 904, // 两个数值相除，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值
	"/":   904,

	"mod": 905, // 两个数值做取模计算，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值
	"%":   905,

	"adds": 921, // 将多个参数进行相加

	"!": 930, // 取反操作符，对于布尔值取反，即true -> false，false -> true。对于其他数值，如果是未定义的变量（即Undefined），返回true，否则返回false

	"not": 931, // 逻辑非操作符，对于布尔值取反，即true -> false，false -> true，对于int、rune、byte等按位取反，即 0 -> 1， 1 -> 0

	"&&": 933, // 逻辑与操作符

	"||": 934, // 逻辑或操作符

	"&": 941, // bit and

	"|": 942, // bit or

	"^":  943, // bit xor
	"&^": 944, // bit and not

	"?":          990, // 三元操作符，用法示例：? $result $a $s1 "abc"，表示判断变量$a中的布尔值，如果为true，则结果为$s1，否则结果值为字符串abc，结果值将放入结果变量result中，如果省略结果参数，结果值将会存入$tmp
	"ifThenElse": 990,

	// "eval":      998, // 计算一个表达式

	"quickEval": 999, // quick eval an expression, use {} to contain an instruction(no nested {} allowed) that return result value in $tmp
	"eval":      999,

	// func related

	"call": 1010, // call a normal function, usage: call $result :func1 $arg1 $arg2...
	// result value could not be omitted, use $drop if not neccessary
	// all arguments/parameters will be put into the local variable "inputL" in the function
	// and the function should return result in local variable "outL"
	// use "ret $result" is a covenient way to set value of $outL and return from the function

	"ret": 1020, // return from a normal function or a fast call function, while for normal function call, can with a paramter for set $outL

	"sealCall": 1050, // new a VM to run a function, output/input through inputG & outG
	// 封装函数，结果参数不可省略，第一个参数可以是代码或编译后、运行上下文，或者起始标号（此时第二个参数应为结束标号），后面参数都将传入行运函数中的$inputG中，出参通过$outG传出

	"runCall": 1055, // 调用称为“行运函数”的代码块，在同一虚拟机、新建的运行上下文中调用函数，结果参数不可省略，第一个参数可以是代码或编译后、运行上下文，或者起始标号（此时第二个参数应为结束标号），后面参数都将传入行运函数中的$inputL中。行运函数会进行函数压栈（以便新的运行上下文中可以deferUpToRoot），入参通过$inputL访问，出参通过$outL传出

	"goRunCall": 1056, // runCall in thread

	"threadCall": 1060, // 并发调用函数，在新虚拟机中运行，函数体内无需返回outG、outL等参数，结果参数不可省略（但仅在调用函数启动线程时如遇错误返回error对象，后续因为是并发调用，返回值无意义），第一个参数如果是个运行上下文对象，后续参数都是传入参数（通过$inputG访问）；第一个参数如果是一个标号或整数，则还需要第二个标号或整数，分别表示并发函数的开始指令标号与结束指令标号；第一个参数也可以是字符串类型的源代码或编译后代码
	"goCall":     1060,

	"go": 1063, // 快速并发调用一个标号处的代码，该段代码应该使用exit命令来表示退出该线程

	"fastCall": 1070, // fast call function, no function stack used, no result value or arguments allowed, use stack or variables for input and output, use fastRet or ret to return

	"fastRet": 1071, // return from fast function, used with fastCall

	// for/range related 循环/遍历相关
	"for": 1080, // for loop, usage: for @`$a < #i10` `++ $a` :cont1 :+1 , if the quick eval result is true(bool value), goto label :cont1, otherwise goto :+1(the next line/instr), the same as in C/C++ "for (; a < 10; a++) {...}"

	"range": 1085, // usage: range 5 :+1 :breakRange1, range #J`[{"a":1,"b":2},{"c":3,"d":4}]` :range1 :+1

	"getIter": 1087, // get i, v or k, v in range

	// array/slice related 数组/切片相关

	"newList":  1101, // 新建一个数组，后接任意个元素作为数组的初始项
	"newArray": 1101,

	"addItem":      1110, //数组中添加项
	"addArrayItem": 1110, //数组中添加项
	"addListItem":  1110,

	"addStrItem": 1111,

	"deleteItem":      1112, //数组中删除项
	"deleteArrayItem": 1112,
	"deleteListItem":  1112,

	"addItems":      1115, // 数组添加另一个数组的值
	"addArrayItems": 1115,
	"addListItems":  1115,

	"getAnyItem": 1120,

	"setAnyItem": 1121,

	"getItem":      1123,
	"getArrayItem": 1123,
	"[]":           1123,

	"setItem":      1124, // 修改数组中某一项的值
	"setArrayItem": 1124,

	"slice": 1130, // 对列表（数组）切片，如果没有指定结果参数，将改变原来的变量。用法示例：slice $list4 $list3 #i1 #i5，将list3进行切片，截取序号1（包含）至序号5（不包含）之间的项，形成一个新的列表，放入变量list4中

	// control related 代码逻辑控制相关
	"continue": 1210, // continue the loop or range, PS "continue 2" means continue the upper loop in nested loop, "continue 1" means continue the upper of upper loop, default is 1 but could be omitted

	"break": 1211, // break the loop or range, PS "break 2" means break the upper loop in nested loop

	// map related 映射相关

	"setMapItem": 1310, // 设置映射项，用法：setMapItem $map1 Name "李白"

	"deleteMapItem": 1312, // 删除映射项

	"getMapItem": 1320, // 获取指定序号的映射项，用法：getMapItem $result $map1 #i2，获取map1中的序号为2的项（即第3项），放入结果变量result中，如果有第4个参数则为默认值（没有找到映射项时使用的值），省略时将是undefined（可与全局内置变量$undefined比较）
	"{}":         1320,

	"getMapKeys": 1331, // 取所有的映射键名，可用于手工遍历等场景

	// object related 对象相关

	"new": 1401, // 新建一个数据或对象，第一个参数为结果放入的变量（不可省略），第二个为字符串格式的数据类型或对象名，后面是可选的0-n个参数，目前支持byte、int等，注意一般获得的结果是引用（或指针）

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

	"quote":   1503, // 将字符串进行转义（加上转义符，如“"”变成“\"”）
	"unquote": 1504, // 将字符串进行解转义

	"isEmpty":     1510, // 判断字符串是否为空
	"是否空串":        1510,
	"isEmptyTrim": 1513, // 判断字符串trim后是否为空

	"strAdd": 1520,

	"strSplit":      1530, // 按指定分割字符串分割字符串，结果参数不可省略，用法示例：strSplit $result $str1 "," 3，其中第3个参数可选（即可省略），表示结果列表最多的项数（例如为3时，将只按逗号分割成3个字符串的列表，后面的逗号将忽略；省略或为-1时将分割出全部）
	"strSplitByLen": 1533, // 按长度拆分一个字符串为数组，注意由于是rune，可能不是按字节长度，例： strSplitByLen $listT $strT 10，可以加第三个参数表示字节数不能超过多少，加第四个参数表示分隔符（遇上分隔符从分隔符后重新计算长度，也就是说分割长度可以超过指定的个数，一般用于有回车的情况）

	"strReplace":   1540, // 字符串替换，用法示例：strReplace $result $str1 $find $replacement
	"strReplaceIn": 1543, // 字符串替换，可同时替换多个子串，用法示例：strReplace $result $str1 $find1 $replacement1 $find2 $replacement2

	"trim":    1550, // 字符串首尾去空白，非字符串将自动转换为字符串
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
	"bytesToHex":  1605, // 以16进制形式输出字节数组
	"bytesToHexX": 1606, // 以16进制形式输出字节数组，字节中间以空格分割

	// thread related 并发/线程相关
	"lock":   1701, // lock an object which is lockable
	"unlock": 1703, // unlock an object which is unlockable

	"lockN":    1721, // lock a global, internal, predefined lock in a lock pool/array, 0 <= N < 10
	"unlockN":  1722, // unlock a global, internal, predefined lock, 0 <= N < 10
	"tryLockN": 1723, // try lock a global, internal, predefined lock, 0 <= N < 10

	"readLockN":    1725, // read lock a global, internal, predefined lock in a lock pool/array, 0 <= N < 10
	"readUnlockN":  1726, // read unlock a global, internal, predefined lock, 0 <= N < 10
	"tryReadLockN": 1727, // try read lock a global, internal, predefined lock, 0 <= N < 10

	// time related 时间相关
	"now": 1910, // 获取当前时间

	"nowStrCompact": 1911, // 获取简化的当前时间字符串，如20220501080930
	"nowStr":        1912, // 获取当前时间字符串的正式表达
	"nowStrFormal":  1912, // 获取当前时间字符串的正式表达
	"nowTick":       1913, // 获取当前时间的Unix时间戳形式
	"timestamp":     1913,

	"nowUTC": 1918,

	"timeSub": 1921, // 时间进行相减操作

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

	"parseCommandLine": 10011, //-> 分析命令行字符串，类似os.Args的获取过程

	// print related
	"pln": 10410, // same as println function in other languages

	"plo": 10411, // print a value with its type

	"plos": 10412, // 输出多个变量或数值的类型和值

	"pr":  10415, // the same as print in other languages
	"prf": 10416, // the same as printf in other languages

	"pl": 10420,

	"plNow": 10422, // 相当于pl，之前多输出一个时间

	"plv": 10430,

	"plvsr": 10433, // 输出多个变量或数值的值的内部表达形式，之间以换行间隔

	"plErr":  10440, // 输出一个error（表示错误的数据类型）信息
	"plErrX": 10441, // 输出一个error（表示错误的数据类型）或TXERROR字符串信息

	"plErrStr": 10450, // 输出一个TXERROR字符串（表示错误的字符串，以TXERROR:开头，后面一般是错误原因描述）信息

	"spr": 10460, // 相当于其它语言的sprintf函数

	// scan/input related 输入相关
	"scanf":  10511, // 相当于其它语言的scanf函数
	"sscanf": 10512, // 相当于其它语言的sscanf函数

	// convert related 转换相关
	"convert": 10810, // 转换数值类型，例如 convert $a int

	"hex":        10821, // 16进制编码，对于数字高位在后
	"hexb":       10822, // 16进制编码，对于数字高位在前
	"unhex":      10823, // 16进制解码，结果是一个字节列表
	"hexToBytes": 10823,
	"toHex":      10824, // 任意数值16进制编码
	"hexToByte":  10825, // 16进制编码转字节

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
	"errStrf":   10915, // 生成TXERROR字符串，用法：errStrf $result "error: %v" $errMsg\
	"getErrStr": 10921, // 获取TXERROR字符串中的错误原因信息（即TXERROR:后的内容）

	"checkErrStr": 10931, // 判断是否是TXERROR字符串，是则退出程序运行

	// error related / err string(with the prefix "TXERROR:" ) related error相关

	"isErr":     10941, // 判断是否是error对象，结果参数不可省略，除结果参数外第一个参数是需要确定是否是error的对象，第二个可选变量是如果是error时，包含的错误描述信息
	"getErrMsg": 10942, // 获取error对象的错误信息

	"isErrX": 10943, // 同时判断是否是error对象或TXERROR字符串，用法：isErrX $result $err1 $errMsg，第三个参数可选（结果参数不可省略），如有会放入错误原因信息

	"checkErrX":  10945, // check if variable is error or err string, and terminate the program if true(检查后续变量或数值是否是error对象或TXERROR字符串，是则输出后中止)
	"getErrStrX": 10947, // 获取error对象或TXERROR字符串中的错误原因信息（即TXERROR:后的内容）

	"errf": 10949, // 生成错误对象，类似printf

	// common related 通用相关
	"clear": 12001, // clear various object, and the object with Close method

	"close": 12003, // 关闭文件等具有Close方法的对象

	// http request/response related HTTP请求相关
	"writeResp":       20110, // 写一个HTTP请求的响应
	"setRespHeader":   20111, // 设置一个HTTP请求的响应头，如setRespHeader $responseG "Content-Type" "text/json; charset=utf-8"
	"writeRespHeader": 20112, // 写一个HTTP请求的响应头状态，如writeRespHeader $responseG #i200
	"getReqHeader":    20113, // 获取一个HTTP请求的请求头信息
	"genJsonResp":     20114, // 生成一个JSON格式的响应字符，用法：genJsonResp $result $requestG "success" "Test passed!"，结果格式类似{"Status":"fail", "Value": "network timeout"}，其中Status字段表示响应处理结果状态，一般只有success和fail两种，分别表示成功和失败，如果失败，Value字段中为失败原因，如果成功，Value中为空或需要返回的信息
	"genResp":         20114,
	"serveFile":       20116,

	"newMux":          20121, // 新建一个HTTP请求处理路由对象，等同于 new mux
	"setMuxHandler":   20122, // 设置HTTP请求路由处理函数，用法：setMuxHandler $muxT "/text1" $arg $text1，其中，text1是字符串形式的处理函数代码，arg是可以传入处理函数的一个参数，处理函数内可通过全局变量inputG来访问，另外还有全局变量requestG表示请求对象，responseG表示响应对象，reqNameG表示请求的子路径，paraMapG表示请求的URL（GET）参数或POST参数映射
	"setMuxStaticDir": 20123, // 设置静态WEB服务的目录，用法示例：setMuxStaticDir $muxT "/static/" "./scripts" ，设置处理路由“/static/”后的URL为静态资源服务，第1个参数为newMux指令创建的路由处理器对象变量，第2个参数是路由路径，第3个参数是对应的本地文件路径，例如：访问 http://127.0.0.1:8080/static/basic.xie，而当前目录是c:\tmp，那么实际上将获得c:\tmp\scripts\basic.xie

	"startHttpServer":  20151, // 启动http服务器，用法示例：startHttpServer $resultT ":80" $muxT ；可以后面加-go参数表示以线程方式启动，此时应注意主线程不要退出，否则服务器线程也会随之退出，可以用无限循环等方式保持运行
	"startHttpsServer": 20153, // 启动https(SSL)服务器，用法示例：startHttpsServer $resultT ":443" $muxT /root/server.crt /root/server.key -go

	// web related WEB相关
	"getWeb":      20210, // 发送一个HTTP网络请求，并获取响应结果（字符串格式），getWeb指令除了第一个参数必须是返回结果的变量，第二个参数是访问的URL，其他所有参数都是可选的，method可以是GET、POST等；encoding用于指定返回信息的编码形式，例如GB2312、GBK、UTF-8等；headers是一个JSON格式的字符串，表示需要加上的自定义的请求头内容键值对；参数中还可以有一个映射类型的变量或值，表示需要POST到服务器的参数，另外可加-bytes参数表示传回字节数组结果，用法示例：getWeb $resultT "http://127.0.0.1:80/xms/xmsApi" -method=POST -encoding=UTF-8 -timeout=15 -headers=`{"Content-Type": "application/json"}` $mapT
	"getWebBytes": 20213, // 与getWeb相同，但获取结果为字节数组

	"downloadFile": 20220, // 下载文件

	"getResource":     20291, // 获取JQuery等常用的脚本或其他内置文本资源，一般用于服务器端提供内置的jquery等脚本嵌入，避免从互联网即时加载，第一个的参数是jquery.min.js等js文件的名称，内置资源中如果含有反引号，将被替换成~~~存储，使用getResource时将被自动替换回反引号
	"getResourceRaw":  20292, // 与getResource作用类似，唯一区别是不将~~~替换回反引号
	"getResourceList": 20293, // 获取可获取的资源名称列表

	// html related HTML相关
	"htmlToText": 20310, // 将HTML转换为字符串，用法示例：htmlToText $result $str1 "flat"，第3个参数开始是可选参数，表示HTML转文本时的选项

	// regex related 正则表达式相关
	"regReplace":       20411,
	"regReplaceAllStr": 20411,

	"regFindAll":   20421, // 获取正则表达式的所有匹配，用法示例：regFindAll $result $str1 $regex1 $group
	"regFind":      20423, // 获取正则表达式的第一个匹配，用法示例：regFind $result $str1 $regex1 $group
	"regFindFirst": 20423,
	"regFindIndex": 20425, // 获取正则表达式的第一个匹配的位置，返回一个整数数组，任意值为-1表示没有找到匹配，用法示例：regFindIndex $result $str1 $regex1

	"regMatch": 20431, // 判断字符串是否完全符合正则表达式，用法示例：regMatch $result "abcab" `a.*b`

	"regContains":   20441, // 判断字符串中是否包含符合正则表达式的子串
	"regContainsIn": 20443, // 判断字符串中是否包含符合任意一个正则表达式的子串
	"regCount":      20445, // 计算字符串中包含符合某个正则表达式的子串个数，用法示例：regCount $result $str1 $regex1

	"regSplit": 20451, // 用正则表达式分割字符串

	"regQuote": 20491, // 将一个普通字符串中涉及正则表达式特殊字符进行转义替换以便用于正则表达式中

	// system related

	"sleep": 20501, // sleep for n seconds(float, 0.001 means 1 millisecond)

	"getClipText": 20511, // 获取剪贴板文本

	"setClipText": 20512, // 设置剪贴板文本

	"getEnv":    20521, // 获取环境变量
	"setEnv":    20522, // 设置环境变量
	"removeEnv": 20523, // 删除环境变量

	"systemCmd":       20601, // 执行一条系统命令，例如： systemCmd "cmd" "/k" "copy a.txt b.txt"
	"openWithDefault": 20603, // 用系统默认的方式打开一个文件，例如： openWithDefault "a.jpg"

	"getOSName": 20901, // 获取操作系统名称，如windows,linux,darwin等
	"getOsName": 20901,

	// file related
	"loadText": 21101, // load text from file
	"saveText": 21103, // 保存文本到指定文件

	"loadBytes": 21105, // 从指定文件载入数据（字节列表）

	"saveBytes": 21106, // 保存数据（字节列表）到指定文件

	"loadBytesLimit": 21107, // 从指定文件载入数据（字节列表），不超过指定字节数

	"appendText": 21111, // 追加文本到指定文件末尾

	"writeStr": 21201, // 写入字符串，可以向文件、字节数组、字符串等写入

	"createFile": 21501, // 新建文件，如果带-return参数，将在成功时返回FILE对象，失败时返回error对象，否则返回error对象，成功为nil，-overwrite有重复文件不会提示。如果需要指定文件标志位等，用openFile指令

	"openFile": 21503, // 打开文件，如果带-readOnly参数，则为只读，-write参数可写，-create参数则无该文件时创建一个，-perm=0777可以指定文件权限标志位

	"openFileForRead": 21505, // 打开一个文件，仅为读取内容使用

	"closeFile": 21507, // 关闭文件

	"readByte": 21521, // 从io.Reader或bufio.Reader读取一个字节

	"readBytesN": 21525, // 从io.Reader或bufio.Reader读取多个字节，第二个参数指定所需读取的字节数

	"writeByte": 21531, // 向io.Writer或bufio.Writer写入1个字节

	"writeBytes": 21533, // 向io.Writer或bufio.Writer写入多个字节

	"flush": 21541, // bufio.Writer等具有缓存的对象清除缓存

	"cmpBinFile": 21601, // 逐个字节比较二进制文件，用法： cmpBinFile $result $file1 $file2 -identical -verbose，如果带有-identical参数，则只比较文件异同（遇上第一个不同的字节就返回布尔值false，全相同则返回布尔值true），不带-identical参数时，将返回一个比较结果对象

	"fileExists":   21701, // 判断文件是否存在
	"ifFileExists": 21701,
	"isDir":        21702, // 判断是否是目录
	"getFileSize":  21703, // 获取文件大小
	"getFileInfo":  21705, // 获取文件信息，返回映射对象，参看genFileList命令

	"removeFile": 21801, // 删除文件，用法：remove $rs $fileNameT -dry
	"renameFile": 21803, // 重命名文件
	"copyFile":   21805, // 复制文件，用法 copyFile $result $fileName1 $fileName2，可带参数-force和-bufferSize=100000等

	// path related

	"genFileList": 21901, // 生成目录中的文件列表，即获取指定目录下的符合条件的所有文件，例：getFileList $result `d:\tmp` "-recursive" "-pattern=*" "-exclusive=*.txt" "-withDir" "-verbose"，另有 -compact 参数将只给出Abs、Size、IsDir三项, -dirOnly参数将只列出目录（不包含文件），列表项对象内容类似：map[Abs:D:\tmpx\test1.gox Ext:.gox IsDir:false Mode:-rw-rw-rw- Name:test1.gox Path:test1.gox Size:353339 Time:20210928091734]
	"getFileList": 21901,

	"joinPath": 21902, // join file paths

	"getCurDir": 21905, // get current working directory
	"setCurDir": 21906, // set current working directory

	"getAppDir":    21907, // get the application directory(where execute-file exists)
	"getConfigDir": 21908, // get application config directory

	"extractFileName": 21910, // 从文件路径中获取文件名部分
	"getFileBase":     21910,
	"extractFileExt":  21911, // 从文件路径中获取文件扩展名（后缀）部分
	"extractFileDir":  21912, // 从文件路径中获取文件目录（路径）部分
	"extractPathRel":  21915, // 从文件路径中获取文件相对路径（根据指定的根路径）

	"ensureMakeDirs": 21921,

	// console related 命令行相关
	"getInput":    22001, // 从命令行获取输入，第一个参数开始是提示字符串，可以类似printf加多个参数，用法：getInput $text1 "请输入%v个数字：" #i2
	"getInputf":   22001,
	"getPassword": 22003, // 从命令行获取密码输入（输入字符不显示），第一个参数是提示字符串

	// json related JSON相关
	"toJson": 22101, // 将对象编码为JSON字符串
	"toJSON": 22101,

	"fromJson": 22102, // 将JSON字符串转换为对象
	"fromJSON": 22102,

	// xml related XML相关
	"toXml": 22201, // 将对象编码为XML字符串
	"toXML": 22201,

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
	"toBase64":     24401,
	"base64Decode": 24403, // Base64解码
	"unbase64":     24403,
	"fromBase64":   24403,

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

	// compress/uncompress related 压缩/解压缩相关
	"compress":   26001, // compress string or byte array to byte array
	"uncompress": 26002,
	"decompress": 26002,

	"compressText":   26011, // compress string to string, may even be longer than original
	"uncompressText": 26012,
	"decompressText": 26012,

	// network relate 网络相关
	"getRandomPort": 27001, // 获取一个可用的socket端口（注意：获取后应尽快使用，否则仍有可能被占用）

	"listen": 27101, // net.Listen
	"accept": 27105, // net.Listener.Accept()

	// database related 数据库相关
	"dbConnect": 32101, // 连接数据库，用法示例：dbConnect $db "sqlite3" `c:\tmpx\test.db`，或dbConnect $db "godror" `user/pass@129.0.9.11:1521/testdb`，结果参数外第一个参数为数据库驱动类型，目前支持sqlite3、mysql、mssql、godror（即oracle）等，第二个参数为连接字串
	"连接数据库":     32101,

	"dbClose": 32102, // 关闭数据库连接
	"关闭数据库":   32102,

	"dbQuery": 32103, // 在指定数据库连接上执行一个查询的SQL语句（一般是select等），返回数组，每行是映射（字段名：字段值），用法示例：dbQuery $rs $db $sql $arg1 $arg2 ...
	"查询数据库":   32103,

	"dbQueryMap": 32104, // 在指定数据库连接上执行一个查询的SQL语句（一般是select等），返回一个映射，以指定的数据库记录字段为键名，对应记录为键值，用法示例：dbQueryMap $rs $db $sql $key $arg1 $arg2 ...
	"查询数据库映射":    32104,

	"dbQueryRecs": 32105, // 在指定数据库连接上执行一个查询的SQL语句（一般是select等），返回二维数组（第一行为字段名），用法示例：dbQueryRecs $rs $db $sql $arg1 $arg2 ...
	"查询数据库记录":     32105,

	"dbQueryCount": 32106, // 在指定数据库连接上执行一个查询的SQL语句，返回一个整数，一般用于查询记录条数或者某个整数字段的值等场景，用法示例：dbQueryCount $rs $db `select count(*) from TABLE1 where FIELD1=:v1 and FIELD2=:v2` $arg1 $arg2 ...

	"dbQueryFloat": 32107, // 在指定数据库连接上执行一个查询的SQL语句，返回一个浮点数，用法示例：dbQueryFloat $rs $db `select PRICE from TABLE1 where FIELD1=:v1 and FIELD2=:v2` $arg1 $arg2 ...

	"dbQueryString": 32108, // 在指定数据库连接上执行一个查询的SQL语句，返回一个字符串，用法示例：dbQueryString $rs $db `select NAME from TABLE1 where FIELD1=:v1 and FIELD2=:v2` $arg1 $arg2 ...

	"dbExec": 32111, // 在指定数据库连接上执行一个有操作的SQL语句（一般是insert、update、delete等），用法示例：dbExec $rs $db $sql $arg1 $arg2 ...
	"执行数据库":  32111,

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
	"leSort":        70049, // 将行文本编辑器缓冲区中的行进行排序，唯一参数（可省略，默认为false）表示是否降序排序，例：leSort $result true
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

	// "compressData":   90201, // 压缩数据，用法：compressData $result $data -method=gzip，压缩方法由-method参数指定，默认为gzip，还支持lzw
	// "decompressData": 90203, // 解压缩数据

	// web GUI related 网页界面相关
	"initWebGUIW":   100001, // 初始化Web图形界面编程环境（Windows下IE11版本），如果没有外嵌式浏览器xiewbr，则将其下载到xie语言目录下
	"initWebGuiW":   100001,
	"updateWebGuiW": 100003, // 强制刷新Web图形界面编程环境（Windows下IE11版本），会将最新的外嵌式浏览器xiewbr下载到xie语言目录下

	"initWebGUIC": 100011, // 初始化Web图形界面编程环境（Windows下CEF版本），如果没有外嵌式浏览器xiecbr及相关库文件，则将其下载到xie语言目录下
	"initWebGuiC": 100011,

	// ssh/sftp/ftp related
	"sshConnect":       200001, // 打开一个SSH连接，用法：sshConnect 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码
	"sshOpen":          200001,
	"sshClose":         200003, // 关闭一个SSH连接
	"sshUpload":        200011, // 通过ssh上传一个文件，用法：sshUpload 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -path=文件路径 -remotePath=远端文件路径，可以加-force参数表示覆盖已有文件
	"sshUploadBytes":   200013, // 通过ssh上传一个二进制内容（字节数组）到文件，用法：sshUpload 结果变量 内容变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -remotePath=远端文件路径，可以加-force参数表示覆盖已有文件；内容变量也可以是一个字符串，将自动转换为字节数组
	"sshDownload":      200021, // 通过ssh下载一个文件，用法：sshDownload 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -path=本地文件路径 -remotePath=远端文件路径，可以加-force参数表示覆盖已有文件
	"sshDownloadBytes": 200023, // 通过ssh下载一个文件，结果为字节数组（或error对象），用法：sshDownloadBytes 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -remotePath=远端文件路径，可以加-force参数表示覆盖已有文件

	// excel related
	"excelNew":    210001, // 新建一个excel文件，用法：excelNew $excelFileT
	"excelOpen":   210003, // 打开一个excel文件，用法：excelOpen $excelFileT `d:\tmp\excel1.xlsx`
	"excelClose":  210005, // 关闭一个excel文件
	"excelSaveAs": 210007, // 保存一个excel文件，用法：excelSave $result $excelFileT `d:\tmp\excel1.xlsx`

	"excelWrite": 210009, // 将excel文件内容写入到可写入源（io.Writer），例如文件、标准输出、网页输出http的response excelWrite $result $excelFileT $writer

	"excelReadSheet": 210101, // 读取已打开的excel文件某一个sheet的内容，返回格式是二维数组，用法：excelReadSheet $result $excelFileT sheet1，最后一个参数可以是字符串类型表示按sheet名称读取，或者是一个整数表示按序号读取
	"excelReadCell":  210103, // 读取指定单元格的内容，返回字符串或错误信息，用法：excelReadCell $result $excelFileT "A1"
	"excelWriteCell": 210105, // 将内容写入到指定单元格，用法：excelWriteCell $result $excelFileT "A1" "abc123"
	"excelSetCell":   210105,

	"excelGetSheetList": 210201, // 获取sheet名字列表，结果是字符串数组

	// mail related

	// "mailNewSender": 220001, // 新建一个邮件发送对象，与 new $result "mailSender" 指令效果类似

	// misc related 杂项相关

	"awsSign": 300101,

	// GUI related 图形界面相关
	"guiInit": 400000, // 初始化GUI环境

	"alert":    400001, // 类似JavaScript中的alert，弹出对话框，显示一个字符串或任意数字、对象的字符串表达
	"guiAlert": 400001,

	"msgBox":      400003, // 类似Delphi、VB中的msgBox，弹出带标题的对话框，显示一个字符串，第一个参数是标题，第二个是字符串
	"showInfo":    400003,
	"guiShowInfo": 400003,

	"showError":    400005, // 弹框显示错误信息
	"guiShowError": 400005,

	"getConfirm":    400011, // 显示信息，获取用户的确认
	"guiGetConfirm": 400011,

	"guiNewWindow": 400031,

	"guiMethod": 410001, // 调用GUI生成的对象的方法
	"guiMt":     410001,
}

// type UndefinedStruct struct {
// 	int
// }

// func (o UndefinedStruct) String() string {
// 	return "undefined"
// }

// var Undefined UndefinedStruct = UndefinedStruct{0}

// type VarRef struct {
// 	Type  int // 1, 3: any value, 7: label, 9xx: predefined variables,
// 	Name  string
// 	Value interface{}
// }

type VarRef struct {
	Ref   int // -99 - invalid, -23 - slice of array/slice, -22 - map item, -21 - array/slice item, -17 - reg, -16 - label, -15 - ref, -12 - unref, -11 - seq, -10 - quickEval, -9 - eval, -8 - pop, -7 - peek, -6 - push, -5 - tmp, -4 - pln, -3 - value only, -2 - drop, -1 - debug, 3 normal vars
	Value interface{}
}

func GetVarRefFromArray(aryA []VarRef, idxT int) VarRef {
	if idxT < 0 || idxT >= len(aryA) {
		return VarRef{Ref: -99, Value: nil}
	}

	return aryA[idxT]
}

type Instr struct {
	Code     int
	Cmd      string
	ParamLen int
	Params   []VarRef
	Line     string
}

type GlobalContext struct {
	SyncMap   tk.SyncMap
	SyncQueue tk.SyncQueue
	SyncStack tk.SyncStack

	SyncSeq tk.Seq

	Vars map[string]interface{}

	VerboseLevel int
}

type FuncContext struct {
	Vars map[string]interface{}

	Tmp interface{}

	// ReturnPointer int

	DeferStack *tk.SimpleStack

	// Layer int

	// Parent *FuncContext
}

func NewFuncContext() *FuncContext {
	rs := &FuncContext{}

	rs.Vars = make(map[string]interface{})

	rs.DeferStack = tk.NewSimpleStack(10, tk.Undefined)

	return rs
}

type CompiledCode struct {
	Labels map[string]int

	Source []string

	CodeList []string

	InstrList []Instr

	CodeSourceMap map[int]int
}

func NewCompiledCode() *CompiledCode {
	ccT := &CompiledCode{}

	ccT.Labels = make(map[string]int, 0)

	ccT.Source = make([]string, 0)

	ccT.CodeList = make([]string, 0)

	ccT.InstrList = make([]Instr, 0)

	ccT.CodeSourceMap = make(map[int]int, 0)

	return ccT
}

func ParseLine(commandA string) ([]string, string, error) {
	var args []string
	var lineT string

	firstT := true

	// state: 1 - start, quotes - 2, arg - 3
	state := 1
	current := ""
	quote := "`"
	escapeNext := false

	command := []rune(commandA)

	for i := 0; i < len(command); i++ {
		c := command[i]

		if escapeNext {
			// if c == 'n' {
			// 	current += string('\n')
			// } else if c == 'r' {
			// 	current += string('\r')
			// } else if c == 't' {
			// 	current += string('\t')
			// } else {
			current += string(c)
			// }
			escapeNext = false
			continue
		}

		if c == '\\' && state == 2 && quote == "\"" {
			current += string(c)
			escapeNext = true
			continue
		}

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
		return []string{}, lineT, fmt.Errorf("unclosed quotes: %v", string(command))
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

func ParseVar(strA string, optsA ...interface{}) VarRef {
	// tk.Pl("parseVar: %#v", strA)
	s1T := strings.TrimSpace(strA)

	if strings.HasPrefix(s1T, "`") && strings.HasSuffix(s1T, "`") {
		s1T = s1T[1 : len(s1T)-1]

		return VarRef{-3, s1T} // value(string)
	} else if strings.HasPrefix(s1T, "'") && strings.HasSuffix(s1T, "'") {
		s1T = s1T[1 : len(s1T)-1]

		return VarRef{-3, s1T} // value(string)
	} else if strings.HasPrefix(s1T, `"`) && strings.HasSuffix(s1T, `"`) { // quoted string
		tmps, errT := strconv.Unquote(s1T)

		if errT != nil {
			return VarRef{-3, s1T}
		}

		return VarRef{-3, tmps} // value(string)
	} else {
		if strings.HasPrefix(s1T, "$") {
			var vv interface{} = nil

			if strings.HasSuffix(s1T, "...") {
				vv = "..."

				s1T = s1T[:len(s1T)-3]
			}

			if s1T == "$drop" {
				return VarRef{-2, nil}
			} else if s1T == "$debug" {
				return VarRef{-1, nil}
			} else if s1T == "$pln" {
				return VarRef{-4, nil}
			} else if s1T == "$pop" {
				return VarRef{-8, vv}
			} else if s1T == "$peek" {
				return VarRef{-7, vv}
			} else if s1T == "$push" {
				return VarRef{-6, nil}
			} else if s1T == "$tmp" {
				return VarRef{-5, vv}
			} else if s1T == "$seq" {
				return VarRef{-11, nil}
			} else {
				return VarRef{3, s1T[1:]}
			}
		} else if strings.HasPrefix(s1T, "&") { // ref
			vNameT := s1T[1:]

			if len(vNameT) < 1 {
				return VarRef{-3, s1T}
			}

			return VarRef{-15, ParseVar(vNameT)}
		} else if strings.HasPrefix(s1T, "*") { // unref
			vNameT := s1T[1:]

			if len(vNameT) < 1 {
				return VarRef{-3, s1T}
			}

			return VarRef{-12, ParseVar(vNameT)}
		} else if strings.HasPrefix(s1T, ":") { // labels
			vNameT := s1T[1:]

			if len(vNameT) < 1 {
				return VarRef{-3, s1T}
			}

			return VarRef{-16, vNameT}
		} else if strings.HasPrefix(s1T, "#") { // values
			if len(s1T) < 2 {
				return VarRef{-3, s1T}
			}

			// remainsT := s1T[2:]

			typeT := s1T[1]

			if typeT == 'i' { // int
				c1T, errT := tk.StrToIntQuick(s1T[2:])

				if errT != nil {
					return VarRef{-3, s1T}
				}

				return VarRef{-3, c1T}
			} else if typeT == 'f' { // float
				c1T, errT := tk.StrToFloat64E(s1T[2:])

				if errT != nil {
					return VarRef{-3, s1T}
				}

				return VarRef{-3, c1T}
			} else if typeT == 'b' { // bool
				return VarRef{-3, tk.ToBool(s1T[2:])}
			} else if typeT == 'y' { // byte
				return VarRef{-3, tk.ToByte(s1T[2:])}
			} else if typeT == 'x' { // byte in hex from
				return VarRef{-3, byte(tk.HexToInt(s1T[2:]))}
			} else if typeT == 'B' { // single rune (same as in Golang, like 'a'), only first character in string is used
				runesT := []rune(s1T[2:])

				if len(runesT) < 1 {
					return VarRef{-3, s1T[2:]}
				}

				return VarRef{-3, runesT[0]}
			} else if typeT == 'r' { // rune
				return VarRef{-3, tk.ToRune(s1T[2:])}
			} else if typeT == 's' { // string
				s1DT := s1T[2:]

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, "'") && strings.HasSuffix(s1DT, "'") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, `"`) && strings.HasSuffix(s1DT, `"`) {
					tmps, errT := strconv.Unquote(s1DT)

					if errT != nil {
						return VarRef{-3, s1DT}
					}

					s1DT = tmps

				}

				return VarRef{-3, s1DT}
				// } else if typeT == '~' { // string, but replace ~~~ to `(back quote)
				// 	s1DT := s1T[2:]

				// 	if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
				// 		s1DT = s1DT[1 : len(s1DT)-1]
				// 	}

				// 	s1DT = strings.ReplaceAll(s1DT, "~~~", "`")

				// 	return VarRef{-3, s1DT}
			} else if typeT == 'e' { // error
				s1DT := s1T[2:]

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, "'") && strings.HasSuffix(s1DT, "'") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, `"`) && strings.HasSuffix(s1DT, `"`) {
					tmps, errT := strconv.Unquote(s1DT)

					if errT != nil {
						return VarRef{-3, s1DT}
					}

					s1DT = tmps

				}

				return VarRef{-3, fmt.Errorf("%v", s1DT)}
			} else if typeT == 't' { // time
				s1DT := s1T[2:]

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, "'") && strings.HasSuffix(s1DT, "'") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, `"`) && strings.HasSuffix(s1DT, `"`) {
					tmps, errT := strconv.Unquote(s1DT)

					if errT != nil {
						return VarRef{-3, s1DT}
					}

					s1DT = tmps

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
			} else if typeT == 'J' { // value from JSON
				var objT interface{}

				s1DT := s1T[2:] // tk.UrlDecode(s1T[2:])

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, "'") && strings.HasSuffix(s1DT, "'") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, `"`) && strings.HasSuffix(s1DT, `"`) {
					tmps, errT := strconv.Unquote(s1DT)

					if errT != nil {
						return VarRef{-3, s1DT}
					}

					s1DT = tmps

				}

				// tk.Plv(s1T[2:])
				// tk.Plv(s1DT)

				errT := json.Unmarshal([]byte(s1DT), &objT)
				// tk.Plv(errT)
				if errT != nil {
					return VarRef{-3, s1T}
				}

				// tk.Plv(listT)
				return VarRef{-3, objT}
			} else if typeT == 'L' { // list/array
				var listT []interface{}

				s1DT := s1T[2:] // tk.UrlDecode(s1T[2:])

				if strings.HasPrefix(s1DT, "`") && strings.HasSuffix(s1DT, "`") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, "'") && strings.HasSuffix(s1DT, "'") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, `"`) && strings.HasSuffix(s1DT, `"`) {
					tmps, errT := strconv.Unquote(s1DT)

					if errT != nil {
						return VarRef{-3, s1DT}
					}

					s1DT = tmps

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
				} else if strings.HasPrefix(s1DT, "'") && strings.HasSuffix(s1DT, "'") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, `"`) && strings.HasSuffix(s1DT, `"`) {
					tmps, errT := strconv.Unquote(s1DT)

					if errT != nil {
						return VarRef{-3, s1DT}
					}

					s1DT = tmps

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
				} else if strings.HasPrefix(s1DT, "'") && strings.HasSuffix(s1DT, "'") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, `"`) && strings.HasSuffix(s1DT, `"`) {
					tmps, errT := strconv.Unquote(s1DT)

					if errT != nil {
						return VarRef{-3, s1DT}
					}

					s1DT = tmps

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
				} else if strings.HasPrefix(s1DT, "'") && strings.HasSuffix(s1DT, "'") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, `"`) && strings.HasSuffix(s1DT, `"`) {
					tmps, errT := strconv.Unquote(s1DT)

					if errT != nil {
						return VarRef{-3, s1DT}
					}

					s1DT = tmps

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
				} else if strings.HasPrefix(s1DT, "'") && strings.HasSuffix(s1DT, "'") {
					s1DT = s1DT[1 : len(s1DT)-1]
				} else if strings.HasPrefix(s1DT, `"`) && strings.HasSuffix(s1DT, `"`) {
					tmps, errT := strconv.Unquote(s1DT)

					if errT != nil {
						return VarRef{-3, s1DT}
					}

					s1DT = tmps

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
			} else if typeT == '#' { // regs

				s1DT := s1T[2:] // tk.UrlDecode(s1T[2:])

				if len(s1DT) < 1 {
					return VarRef{-3, s1T}
				}

				return VarRef{-17, tk.ToInt(s1T, 0)}
			}

			return VarRef{-3, s1T}
		} else if strings.HasPrefix(s1T, "@") { // quickEval
			if len(s1T) < 2 {
				return VarRef{-3, s1T}
			}

			s1T = strings.TrimSpace(s1T[1:])

			if strings.HasPrefix(s1T, "`") && strings.HasSuffix(s1T, "`") {
				s1T = s1T[1 : len(s1T)-1]

				return VarRef{-10, s1T} // quick eval value
			} else if strings.HasPrefix(s1T, "'") && strings.HasSuffix(s1T, "'") {
				s1T = s1T[1 : len(s1T)-1]

				return VarRef{-10, s1T} // quick eval value
			} else if strings.HasPrefix(s1T, `"`) && strings.HasSuffix(s1T, `"`) {
				tmps, errT := strconv.Unquote(s1T)

				if errT != nil {
					return VarRef{-10, s1T}
				}

				return VarRef{-10, tmps}
			}

			return VarRef{-10, s1T}
			// } else if strings.HasPrefix(s1T, "^") { // regs
			// 	if len(s1T) < 2 {
			// 		return VarRef{-3, s1T}
			// 	}

			// 	s1T = strings.TrimSpace(s1T[1:])

			// 	return VarRef{-17, tk.ToInt(s1T, 0)}
		} else if strings.HasPrefix(s1T, "%") { // compiled
			if len(s1T) < 2 {
				return VarRef{-3, s1T}
			}

			s1T = strings.TrimSpace(s1T[1:])

			if strings.HasPrefix(s1T, "`") && strings.HasSuffix(s1T, "`") {
				s1T = s1T[1 : len(s1T)-1]

				nc := Compile(s1T)

				// if tk.IsError(nc) {
				// 	return VarRef{-3, nc}
				// }

				return VarRef{-3, nc} // compiled
			} else if strings.HasPrefix(s1T, "'") && strings.HasSuffix(s1T, "'") {
				s1T = s1T[1 : len(s1T)-1]

				nc := Compile(s1T)

				return VarRef{-3, nc} // compiled
			} else if strings.HasPrefix(s1T, `"`) && strings.HasSuffix(s1T, `"`) {
				tmps, errT := strconv.Unquote(s1T)

				if errT != nil {
					return VarRef{-3, errT}
				}

				nc := Compile(tmps)

				return VarRef{-3, nc} // compiled
			}

			return VarRef{-3, Compile(s1T)}
		} else if strings.HasPrefix(s1T, "[") && strings.HasSuffix(s1T, "]") { // array/slice item
			if len(s1T) < 3 {
				return VarRef{-3, s1T}
			}

			s1aT := strings.TrimSpace(s1T[1 : len(s1T)-1])

			listT := strings.Split(s1aT, ",")

			len2T := len(listT)

			if len2T >= 3 { // slice of array/slice/string
				vT := ParseVar(listT[0])

				itemKeyT := listT[1]

				if strings.HasPrefix(itemKeyT, "`") && strings.HasSuffix(itemKeyT, "`") {
					itemKeyT = itemKeyT[1 : len(itemKeyT)-1]
				} else if strings.HasPrefix(itemKeyT, "'") && strings.HasSuffix(itemKeyT, "'") {
					itemKeyT = itemKeyT[1 : len(itemKeyT)-1]
				} else if strings.HasPrefix(itemKeyT, `"`) && strings.HasSuffix(itemKeyT, `"`) {
					tmps, errT := strconv.Unquote(itemKeyT)

					if errT != nil {
						itemKeyT = tmps
					}
				}

				itemKeyEndT := listT[2]

				if strings.HasPrefix(itemKeyEndT, "`") && strings.HasSuffix(itemKeyEndT, "`") {
					itemKeyEndT = itemKeyEndT[1 : len(itemKeyEndT)-1]
				} else if strings.HasPrefix(itemKeyEndT, "'") && strings.HasSuffix(itemKeyEndT, "'") {
					itemKeyEndT = itemKeyEndT[1 : len(itemKeyEndT)-1]
				} else if strings.HasPrefix(itemKeyEndT, `"`) && strings.HasSuffix(itemKeyEndT, `"`) {
					tmps, errT := strconv.Unquote(itemKeyEndT)

					if errT != nil {
						itemKeyEndT = tmps
					}
				}

				return VarRef{-23, []interface{}{vT, ParseVar(itemKeyT), ParseVar(itemKeyEndT)}}

			}

			if len2T < 2 {
				listT = strings.SplitN(s1aT, "|", 2)

				if len(listT) < 2 {
					return VarRef{-3, s1T}
				}
			}

			vT := ParseVar(listT[0])

			itemKeyT := listT[1]

			if strings.HasPrefix(itemKeyT, "`") && strings.HasSuffix(itemKeyT, "`") {
				itemKeyT = itemKeyT[1 : len(itemKeyT)-1]
			} else if strings.HasPrefix(itemKeyT, "'") && strings.HasSuffix(itemKeyT, "'") {
				itemKeyT = itemKeyT[1 : len(itemKeyT)-1]
			} else if strings.HasPrefix(itemKeyT, `"`) && strings.HasSuffix(itemKeyT, `"`) {
				tmps, errT := strconv.Unquote(itemKeyT)

				if errT != nil {
					itemKeyT = tmps
				}
			}

			return VarRef{-21, []interface{}{vT, ParseVar(itemKeyT)}}
		} else if strings.HasPrefix(s1T, "{") && strings.HasSuffix(s1T, "}") { // map item
			if len(s1T) < 3 {
				return VarRef{-3, s1T}
			}

			s1aT := strings.TrimSpace(s1T[1 : len(s1T)-1])

			listT := strings.SplitN(s1aT, ",", 2)

			if len(listT) < 2 {
				listT = strings.SplitN(s1aT, "|", 2)

				if len(listT) < 2 {
					return VarRef{-3, s1T}
				}
			}

			vT := ParseVar(listT[0])

			itemKeyT := listT[1]

			if strings.HasPrefix(itemKeyT, "`") && strings.HasSuffix(itemKeyT, "`") {
				itemKeyT = itemKeyT[1 : len(itemKeyT)-1]
			} else if strings.HasPrefix(itemKeyT, "'") && strings.HasSuffix(itemKeyT, "'") {
				itemKeyT = itemKeyT[1 : len(itemKeyT)-1]
			} else if strings.HasPrefix(itemKeyT, `"`) && strings.HasSuffix(itemKeyT, `"`) {
				tmps, errT := strconv.Unquote(itemKeyT)

				if errT != nil {
					itemKeyT = tmps
				}
			}

			return VarRef{-22, []interface{}{vT, ParseVar(itemKeyT)}}
		}
	}

	return VarRef{-3, s1T}
}

func Compile(codeA string) interface{} {
	// codeA = strings.ReplaceAll(codeA, "~~~", "`")

	p := NewCompiledCode()

	originCodeLenT := 0

	sourceT := tk.SplitLines(codeA)

	p.Source = append(p.Source, sourceT...)

	pointerT := originCodeLenT

	for i := 0; i < len(sourceT); i++ {
		v := strings.TrimSpace(sourceT[i])

		if tk.StartsWith(v, "//") || tk.StartsWith(v, "#") {
			continue
		}

		if tk.StartsWith(v, ":") {
			labelT := strings.TrimSpace(v[1:])

			_, ok := p.Labels[labelT]

			if !ok {
				p.Labels[labelT] = pointerT
			} else {
				return fmt.Errorf("编译错误(行 %v %v): 重复的标号", i+1, tk.LimitString(p.Source[i], 50))
			}

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
					return fmt.Errorf("代码解析错误: ` 未成对(%v)", i)
				}

				i = j
			}
		}

		v = strings.TrimSpace(v)

		if v == "" {
			continue
		}

		p.CodeList = append(p.CodeList, v)
		p.CodeSourceMap[pointerT] = originCodeLenT + iFirstT
		pointerT++
	}

	for i := originCodeLenT; i < len(p.CodeList); i++ {
		// listT := strings.SplitN(v, " ", 3)
		v := p.CodeList[i]

		listT, lineT, errT := ParseLine(v)
		if errT != nil {
			return fmt.Errorf("参数解析失败：%v", errT)
		}

		lenT := len(listT)

		instrNameT := strings.TrimSpace(listT[0])

		codeT, ok := InstrNameSet[instrNameT]

		if !ok {
			instrT := Instr{Code: codeT, Cmd: instrNameT, ParamLen: 1, Params: []VarRef{VarRef{Ref: -3, Value: v}}, Line: lineT} //&([]VarRef{})}
			p.InstrList = append(p.InstrList, instrT)

			return fmt.Errorf("编译错误(行 %v/%v %v): 未知指令", i, p.CodeSourceMap[i]+1, tk.LimitString(p.Source[p.CodeSourceMap[i]], 50))
		}

		// if codeT == 109 { // defer
		// 	if lenT < 2 {
		// 		instrT := Instr{Code: codeT, Cmd: instrNameT, ParamLen: 1, Params: []VarRef{VarRef{Ref: -3, Value: v}}, Line: lineT} //&([]VarRef{})}
		// 		p.InstrList = append(p.InstrList, instrT)

		// 		return fmt.Errorf("编译错误(行 %v/%v %v): 空的defer指令", i, p.CodeSourceMap[i]+1, tk.LimitString(p.Source[p.CodeSourceMap[i]], 50))
		// 	}

		// 	deferCmdT := strings.TrimSpace(listT[1])

		// 	if strings.HasPrefix(deferCmdT, "$") {
		// 		// 157 runPiece
		// 		instrT := Instr{Code: 157, Cmd: "runPiece", Params: []VarRef{VarRef{Ref: -5, Value: nil}, ParseVar(deferCmdT, i)}, ParamLen: 2, Line: lineT} //&([]VarRef{})}

		// 		p.InstrList = append(p.InstrList, instrT)

		// 		continue

		// 	} else {
		// 		deferCodeT, ok := InstrNameSet[deferCmdT]

		// 		if !ok {
		// 			instrT := Instr{Code: deferCodeT, Cmd: deferCmdT, ParamLen: 1, Params: []VarRef{VarRef{Ref: -3, Value: v}}, Line: lineT} //&([]VarRef{})}
		// 			p.InstrList = append(p.InstrList, instrT)

		// 			return fmt.Errorf("编译错误(行 %v/%v %v): 未知的defer指令", i, p.CodeSourceMap[i]+1, tk.LimitString(p.Source[p.CodeSourceMap[i]], 50))

		// 		}

		// 		instrDeferT := Instr{Code: deferCodeT, Cmd: deferCmdT, Params: make([]VarRef, 0, lenT-2), ParamLen: lenT - 2, Line: tk.RemoveFirstSubString(strings.TrimSpace(lineT), deferCmdT)} //&([]VarRef{})}

		// 		list3T := []VarRef{}

		// 		for j, jv := range listT {
		// 			if j < 2 {
		// 				continue
		// 			}

		// 			list3T = append(list3T, ParseVar(jv, i))
		// 		}

		// 		instrDeferT.Params = append(instrDeferT.Params, list3T...)
		// 		instrDeferT.ParamLen = lenT - 2

		// 		instrT := Instr{Code: codeT, Cmd: instrNameT, Params: []VarRef{VarRef{Ref: -3, Value: instrDeferT}}, ParamLen: 1, Line: lineT} //&([]VarRef{})}

		// 		p.InstrList = append(p.InstrList, instrT)

		// 		continue

		// 	}

		// }

		instrT := Instr{Code: codeT, Cmd: instrNameT, Params: make([]VarRef, 0, lenT-1), Line: lineT} //&([]VarRef{})}

		list3T := []VarRef{}

		for j, jv := range listT {
			if j == 0 {
				continue
			}

			list3T = append(list3T, ParseVar(jv, i))
		}

		instrT.Params = append(instrT.Params, list3T...)
		instrT.ParamLen = lenT - 1

		p.InstrList = append(p.InstrList, instrT)
	}

	// tk.Plv(p.SourceM)
	// tk.Plv(p.CodeListM)
	// tk.Plv(p.CodeSourceMapM)

	return p
}

type RunningContext struct {
	Labels map[string]int

	Source []string

	CodeList []string

	InstrList []Instr

	CodeSourceMap map[int]int

	CodePointer int

	PointerStack *tk.SimpleStack

	FuncStack *tk.SimpleStack

	ErrorHandler int

	// Parent interface{}
}

func (p *RunningContext) Initialize() {
	p.Labels = make(map[string]int)
	p.Source = make([]string, 0)
	p.CodeList = make([]string, 0)
	p.InstrList = make([]Instr, 0)
	p.CodeSourceMap = make(map[int]int)
	p.CodePointer = 0

	p.PointerStack = tk.NewSimpleStack(10, tk.Undefined)

	p.FuncStack = tk.NewSimpleStack(10, tk.Undefined)

	p.ErrorHandler = -1
}

func (p *RunningContext) LoadCompiled(compiledA *CompiledCode) error {
	// tk.Plv(compiledA)

	// load labels
	originalLenT := len(p.CodeList)
	originalSourceLenT := len(p.Source)

	for k, _ := range compiledA.Labels {
		_, ok := p.Labels[k]

		if ok {
			return fmt.Errorf("duplicate label: %v", k)
		}
	}

	for k, v := range compiledA.Labels {
		p.Labels[k] = originalLenT + v
	}

	// load codeList, instrList
	p.Source = append(p.Source, compiledA.Source...)

	p.CodeList = append(p.CodeList, compiledA.CodeList...)
	// for _, v := range compiledA.CodeList {
	// 	p.CodeList = append(p.CodeList, v)
	// }

	p.InstrList = append(p.InstrList, compiledA.InstrList...)
	// for _, v := range compiledA.InstrList {
	// 	p.InstrList = append(p.InstrList, v)
	// }

	for k, v := range compiledA.CodeSourceMap {
		p.CodeSourceMap[originalLenT+k] = originalSourceLenT + v
	}

	return nil
}

func (p *RunningContext) Extract(startA, endA int) interface{} {
	newRunT := NewRunningContext().(*RunningContext)

	originalLenT := len(p.CodeList)

	if startA < 0 || startA >= originalLenT {
		return fmt.Errorf("start index not in range: %v(%v)", startA, originalLenT)
	}

	if endA < 0 || endA >= originalLenT {
		return fmt.Errorf("end index not in range: %v(%v)", endA, originalLenT)
	}

	if startA > endA {
		return fmt.Errorf("startA > endA: %v(%v)", startA, endA)
	}

	newLenT := endA - startA + 1

	// load labels

	for k, v := range p.Labels {
		if v >= startA && v <= endA {
			newRunT.Labels[k] = v - startA
		}
	}

	// load codeList, instrList

	startSourceA := p.CodeSourceMap[startA]
	endSourceA := p.CodeSourceMap[endA]

	// tk.Pl("s: %v e: %v", startSourceA, endSourceA)

	for i := startSourceA; i <= endSourceA; i++ {
		newRunT.Source = append(newRunT.Source, p.Source[i])
	}

	for i := 0; i < newLenT; i++ {
		newRunT.CodeList = append(newRunT.CodeList, p.CodeList[startA+i])
		newRunT.InstrList = append(newRunT.InstrList, p.InstrList[startA+i])

		newRunT.CodeSourceMap[i] = p.CodeSourceMap[startA+i] - startSourceA
	}

	newRunT.CodePointer = 0

	return newRunT
}

func (p *RunningContext) ExtractCompiled(startA, endA int) interface{} {
	newT := NewCompiledCode()

	originalLenT := len(p.CodeList)

	if startA < 0 || startA >= originalLenT {
		return fmt.Errorf("start index not in range: %v(%v)", startA, originalLenT)
	}

	if endA < 0 || endA >= originalLenT {
		return fmt.Errorf("end index not in range: %v(%v)", endA, originalLenT)
	}

	if startA > endA {
		return fmt.Errorf("startA > endA: %v(%v)", startA, endA)
	}

	newLenT := endA - startA + 1

	// load labels

	for k, v := range p.Labels {
		if v >= startA && v <= endA {
			newT.Labels[k] = v - startA
		}
	}

	// load codeList, instrList

	startSourceA := p.CodeSourceMap[startA]
	endSourceA := p.CodeSourceMap[endA]

	// tk.Pl("s: %v e: %v", startSourceA, endSourceA)

	for i := startSourceA; i <= endSourceA; i++ {
		newT.Source = append(newT.Source, p.Source[i])
	}

	for i := 0; i < newLenT; i++ {
		newT.CodeList = append(newT.CodeList, p.CodeList[startA+i])
		newT.InstrList = append(newT.InstrList, p.InstrList[startA+i])

		newT.CodeSourceMap[i] = p.CodeSourceMap[startA+i] - startSourceA
	}

	return newT
}

func (p *RunningContext) LoadCode(codeA string) error {
	rs := Compile(codeA)

	if e1, ok := rs.(error); ok {
		return e1
	}

	e2 := p.LoadCompiled(rs.(*CompiledCode))

	if e2 != nil {
		return e2
	}

	return nil
}

func (p *RunningContext) Import(newRunA *RunningContext) error {
	if newRunA == nil {
		return nil
	}

	originalLenT := len(p.CodeList)

	originalSourceLenT := len(p.Source)

	for k, _ := range newRunA.Labels {
		_, ok := p.Labels[k]

		if ok {
			return fmt.Errorf("duplicate label: %v", k)
		}
	}

	for k, v := range newRunA.Labels {
		p.Labels[k] = originalLenT + v
	}

	p.Source = append(p.Source, newRunA.Source...)

	p.CodeList = append(p.CodeList, newRunA.CodeList...)
	// for _, v := range compiledA.CodeList {
	// 	p.CodeList = append(p.CodeList, v)
	// }

	p.InstrList = append(p.InstrList, newRunA.InstrList...)
	// for _, v := range compiledA.InstrList {
	// 	p.InstrList = append(p.InstrList, v)
	// }

	for k, v := range newRunA.CodeSourceMap {
		p.CodeSourceMap[originalLenT+k] = originalSourceLenT + v
	}

	return nil
}

func (p *RunningContext) Load(inputA ...interface{}) error {

	var inputT interface{} = nil

	if len(inputA) > 0 {
		inputT = inputA[0]
	}

	if inputT == nil {
		return nil
	}

	c1, ok := inputT.(*CompiledCode)

	if ok {
		errT := p.LoadCompiled(c1)

		if errT != nil {
			return errT
		}

		return nil
	}

	s1, ok := inputT.(string)
	if ok {
		e2 := p.LoadCode(s1)

		if e2 != nil {
			return e2
		}

		return nil
	}

	return fmt.Errorf("invalid parameter: %#v", inputT)
}

// return <0 as error
func (p *RunningContext) GetLabelIndex(inputA interface{}) int {
	c, ok := inputA.(int)

	if ok {
		return c
	}

	s2 := tk.ToStr(inputA)

	if strings.HasPrefix(s2, ":") {
		s2 = s2[1:]
	}

	if len(s2) > 1 {
		if strings.HasPrefix(s2, "+") {
			return p.CodePointer + tk.ToInt(s2[1:])
		} else if strings.HasPrefix(s2, "-") {
			return p.CodePointer - tk.ToInt(s2[1:])
		} else {
			labelPointerT, ok := p.Labels[s2]

			if ok {
				return labelPointerT
			}
		}
	}

	return -1
}

func (p *RunningContext) GetFuncContext(layerT int) *FuncContext {
	if layerT < 0 {
		return p.FuncStack.Peek().(*FuncContext)
	}

	return p.FuncStack.PeekLayer(layerT).(*FuncContext)
}

func (p *RunningContext) GetCurrentFuncContext() *FuncContext {
	return p.FuncStack.Peek().(*FuncContext)
}

func (p *FuncContext) RunDefer(vmA *XieVM, rcA *RunningContext) error {
	for {
		instrT := p.DeferStack.Pop()

		// tk.Pl("\nDeferStack.Pop: %#v\n", instrT)

		if instrT == nil || tk.IsUndefined(instrT) {
			break
		}

		nc, ok := instrT.(*CompiledCode)

		if ok {
			if GlobalsG.VerboseLevel > 1 {
				tk.Pl("defer run: %#v", nc)
			}

			// tk.Plo("code piece", p.DeferStack.Size(), nc)

			rs := QuickRunPiece(vmA, rcA, nc)

			if tk.IsError(rs) {
				return fmt.Errorf("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStrX(rs))
			}

			continue
		}

		nv, ok := instrT.(*Instr)

		if !ok {
			nvv, ok := instrT.(Instr)

			if ok {
				nv = &nvv
			} else {
				return fmt.Errorf("invalid instruction: %#v", instrT)
			}
		}

		if GlobalsG.VerboseLevel > 1 {
			tk.Pl("defer run: %v", nv)
		}

		rs := RunInstr(vmA, rcA, nv)

		if tk.IsError(rs) {
			return fmt.Errorf("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStrX(rs))
		}
	}

	return nil
}

func (p *XieVM) RunDefer(runA *RunningContext) error {
	if runA == nil {
		runA = p.Running
	}

	var currentFuncT *FuncContext

	lenT := runA.FuncStack.Size()

	if lenT < 1 {
		currentFuncT = p.RootFunc
	}

	rs := currentFuncT.RunDefer(p, runA)

	return rs
}

func (p *XieVM) RunDeferUpToRoot(runA *RunningContext) error {
	if runA == nil {
		runA = p.Running
	}

	lenT := runA.FuncStack.Size()

	if lenT > 0 {
		rs := runA.RunDeferUpToRoot(p)

		if tk.IsError(rs) {
			return rs
		}
	}

	rs := p.RootFunc.RunDefer(p, runA)

	return rs
}

func (p *RunningContext) RunDeferUpToRoot(vmA *XieVM) error {
	// if p.Parent == nil {
	// 	return fmt.Errorf("no parent VM: %v", p.Parent)
	// }

	lenT := p.FuncStack.Size()

	if lenT < 1 {
		return nil
	}

	for i := lenT - 1; i >= 0; i-- {
		contextT := p.FuncStack.PeekLayer(i).(*FuncContext)

		rs := contextT.RunDefer(vmA, p)

		if tk.IsError(rs) {
			return rs
		}
	}

	return nil

}

func GetDeferStack(vmA *XieVM, rcA *RunningContext) *tk.SimpleStack {
	stackT := tk.NewSimpleStack(0, tk.Undefined)

	lenT := rcA.FuncStack.Size()

	for i := lenT - 1; i >= 0; i-- {
		contextT := rcA.FuncStack.PeekLayer(i).(*FuncContext)

		len1T := contextT.DeferStack.Size()
		for j := len1T - 1; j >= 0; j-- {
			deferT := contextT.DeferStack.PeekLayer(j)

			if tk.IsUndefined(deferT) {
				break
			}

			stackT.Push(deferT)

		}
	}

	len2T := vmA.RootFunc.DeferStack.Size()
	for j := len2T - 1; j >= 0; j-- {
		deferT := vmA.RootFunc.DeferStack.PeekLayer(j)

		if tk.IsUndefined(deferT) {
			break
		}

		stackT.Push(deferT)

	}

	return stackT.Reverse()
}

func RunDefer(vmA *XieVM, runA *RunningContext) error {
	lenT := runA.FuncStack.Size()

	if lenT < 1 {
		return nil
	}

	contextT := runA.FuncStack.Peek().(*FuncContext)

	rs := contextT.RunDefer(vmA, runA)

	if tk.IsError(rs) {
		return rs
	}

	return nil

}

func NewRunningContext(inputA ...interface{}) interface{} {
	var inputT interface{} = nil

	if len(inputA) > 0 {
		inputT = inputA[0]
	}

	nvr, ok := inputT.(*RunningContext)

	if ok {
		return nvr
	}

	rc := &RunningContext{}

	rc.Initialize()

	if inputT == nil {
		return rc
	}

	c1, ok := inputT.(*CompiledCode)

	if ok {
		errT := rc.LoadCompiled(c1)

		if errT != nil {
			return errT
		}

		return rc
	}

	s1, ok := inputT.(string)
	if ok {
		e2 := rc.LoadCode(s1)

		if e2 != nil {
			return e2
		}

		return rc
	}

	return fmt.Errorf("invalid parameter: %#v", inputT)
}

type XieVM struct {
	Regs  []interface{}
	Stack *tk.SimpleStack

	// Vars map[string]interface{}

	RootFunc *FuncContext

	Running *RunningContext
}

// inputA for a RunningContext
func NewVM(inputA ...interface{}) interface{} {
	var inputT interface{} = nil

	if len(inputA) > 0 {
		inputT = inputA[0]
	}

	rs := &XieVM{}

	rs.Regs = make([]interface{}, 30)
	rs.Stack = tk.NewSimpleStack(10, tk.Undefined)

	rs.RootFunc = NewFuncContext()
	// rs.Vars = make(map[string]interface{}, 0)

	var runningT interface{}

	if inputT != nil {
		runningT = NewRunningContext(inputT)
	} else {
		runningT = NewRunningContext()
	}

	if tk.IsError(runningT) {
		return runningT
	}

	rs.Running = runningT.(*RunningContext)

	// rs.Running.Parent = rs

	// rs.FuncStack = tk.NewSimpleStack(10, tk.Undefined)

	// rs.Running.FuncStack.Push(NewFuncContext())

	// set global variables
	rs.SetVar(rs.Running, "backQuoteG", "`")
	rs.SetVar(rs.Running, "undefinedG", tk.Undefined)
	rs.SetVar(rs.Running, "nilG", nil)
	rs.SetVar(rs.Running, "newLineG", "\n")
	// rs.SetVar("tmp", "")

	return rs
}

func NewVMQuick(inputA ...interface{}) *XieVM {
	rs := NewVM(inputA...)

	if tk.IsError(rs) {
		return nil
	}

	return rs.(*XieVM)
}

func (p *XieVM) GetCodeLen(runA *RunningContext) int {
	if runA == nil {
		runA = p.Running
	}

	return len(runA.CodeList)
}

func (p *XieVM) GetSwitchVarValue(runA *RunningContext, argsA []string, switchStrA string, defaultA ...string) string {
	if runA == nil {
		runA = p.Running
	}

	vT := tk.GetSwitch(argsA, switchStrA, defaultA...)

	vr := ParseVar(vT)

	return tk.ToStr(p.GetVarValue(runA, vr))
}

func (p *XieVM) GetSwitchVarValueI(runA *RunningContext, argsA []interface{}, switchStrA string, defaultA ...string) string {
	if runA == nil {
		runA = p.Running
	}

	vT := tk.GetSwitchI(argsA, switchStrA, defaultA...)

	vr := ParseVar(vT)

	return tk.ToStr(p.GetVarValue(runA, vr))
}

func (p *XieVM) GetFuncContext(runA *RunningContext, layerT int) *FuncContext {
	if runA == nil {
		runA = p.Running
	}

	if runA.FuncStack.Size() < 1 {
		return p.RootFunc
	}

	if layerT < 0 {
		return runA.FuncStack.Peek().(*FuncContext)
	}

	if layerT == 0 {
		return p.RootFunc
	}

	if layerT-1 >= 0 && layerT-1 < runA.FuncStack.Size() {
		return runA.FuncStack.PeekLayer(layerT - 1).(*FuncContext)
	}

	return nil
}

func (p *XieVM) GetCurrentFuncContext(runA *RunningContext) *FuncContext {
	if runA == nil {
		runA = p.Running
	}

	if runA.FuncStack.Size() < 1 {
		return p.RootFunc
	}

	return runA.FuncStack.Peek().(*FuncContext)
}

func (p *RunningContext) PushFunc() {
	funcContextT := NewFuncContext()

	// funcContextT.ReturnPointer = p.CodePointer + 1
	// funcContextT.Layer = p.FuncStack.Size() + 1

	// &FuncContext{Vars: make(map[string]interface{}, 0), ReturnPointer: p.CodePointer + 1, Layer: p.FuncStack.Size(), DeferStack: tk.NewSimpleStack(10, tk.Undefined)} // , Parent: p.FuncStack.Peek().(*FuncContext)

	p.FuncStack.Push(funcContextT)
}

func (p *RunningContext) PopFunc() error {

	funcContextItemT := p.FuncStack.Pop()

	if tk.IsUndefined(funcContextItemT) {
		return fmt.Errorf("no function in func stack")
	}

	// funcContextT := funcContextItemT.(*FuncContext)

	return nil
}

func GetVarRefInParams(varRefsA []VarRef, idxA int) VarRef {
	if idxA < 0 || idxA >= len(varRefsA) {
		return VarRef{Ref: -99, Value: nil}
	}

	return varRefsA[idxA]
}

func (p *XieVM) GetVarValue(runA *RunningContext, vA VarRef) interface{} {
	if runA == nil {
		runA = p.Running
	}

	idxT := vA.Ref

	if idxT == -2 {
		return tk.Undefined
	}

	if idxT == -3 {
		return vA.Value
	}

	if idxT == -5 {
		return p.GetCurrentFuncContext(runA).Tmp
	}

	if idxT == -11 {
		return GlobalsG.SyncSeq.Get()
	}

	if idxT == -8 {
		return p.Stack.Pop()
	}

	if idxT == -7 {
		return p.Stack.Peek()
	}

	if idxT == -1 { // $debug
		return tk.ToJSONX(p, "-indent", "-sort")
	}

	if idxT == -6 {
		return tk.Undefined
	}

	if idxT == -10 {
		// tk.Pln("getvarvalue", vA.Value)
		return QuickEval(tk.ToStr(vA.Value), p, runA)
	}

	if idxT == -16 { // labels
		return runA.GetLabelIndex(vA.Value)
	}

	if idxT == -17 { // regs
		return p.Regs[tk.ToInt(vA.Value, 0)]
	}

	if idxT == -21 { // array/slice item
		nv := vA.Value.([]interface{})
		return tk.GetArrayItem(p.GetVarValue(runA, nv[0].(VarRef)), tk.ToInt(p.GetVarValue(runA, nv[1].(VarRef)), 0))
	}

	if idxT == -22 { // map item
		nv := vA.Value.([]interface{})
		return tk.GetMapItem(p.GetVarValue(runA, nv[0].(VarRef)), p.GetVarValue(runA, nv[1].(VarRef)))
	}

	if idxT == -23 { // slice of array/slice
		nv := vA.Value.([]interface{})
		return tk.GetArraySlice(p.GetVarValue(runA, nv[0].(VarRef)), tk.ToInt(p.GetVarValue(runA, nv[1].(VarRef)), 0), tk.ToInt(p.GetVarValue(runA, nv[2].(VarRef)), 0))
	}

	if idxT == -12 { // unref
		rs, errT := tk.GetRefValue(p.GetVarValue(runA, vA.Value.(VarRef)))

		if errT != nil {
			return tk.Undefined
		}

		return rs
	}

	if idxT == -15 { // ref
		return tk.Undefined
	}

	if idxT == 3 { // normal variables
		lenT := runA.FuncStack.Size()

		for idxT := lenT - 1; idxT >= 0; idxT-- {
			loopFunc := runA.FuncStack.PeekLayer(idxT).(*FuncContext)
			nv, ok := loopFunc.Vars[vA.Value.(string)]

			if ok {
				return nv
			}
		}

		nv, ok := p.RootFunc.Vars[vA.Value.(string)]

		if ok {
			return nv
		}

		return tk.Undefined

	}

	return tk.Undefined

}

func (p *XieVM) GetVarValueGlobal(runA *RunningContext, vA VarRef) interface{} {
	if runA == nil {
		runA = p.Running
	}

	idxT := vA.Ref

	if idxT == -2 {
		return tk.Undefined
	}

	if idxT == -3 {
		return vA.Value
	}

	if idxT == -5 {
		return p.RootFunc.Tmp
	}

	if idxT == -11 {
		return GlobalsG.SyncSeq.Get()
	}

	if idxT == -8 {
		return p.Stack.Pop()
	}

	if idxT == -7 {
		return p.Stack.Peek()
	}

	if idxT == -1 { // $debug
		return tk.ToJSONX(p, "-indent", "-sort")
	}

	if idxT == -6 {
		return tk.Undefined
	}

	if idxT == -10 {
		// tk.Pln("getvarvalue", vA.Value)
		return QuickEval(tk.ToStr(vA.Value), p, runA)
	}

	if idxT == -16 { // labels
		return runA.GetLabelIndex(vA.Value)
	}

	if idxT == -17 { // regs
		return p.Regs[tk.ToInt(vA.Value, 0)]
		return nil
	}

	if idxT == -21 { // array/slice item
		nv := vA.Value.([]interface{})
		return tk.GetArrayItem(p.GetVarValue(runA, nv[0].(VarRef)), tk.ToInt(p.GetVarValue(runA, nv[1].(VarRef)), 0))
	}

	if idxT == -22 { // map item
		nv := vA.Value.([]interface{})
		return tk.GetMapItem(p.GetVarValue(runA, nv[0].(VarRef)), p.GetVarValue(runA, nv[1].(VarRef)))
	}

	if idxT == -23 { // slice of array/slice
		nv := vA.Value.([]interface{})
		return tk.GetArraySlice(p.GetVarValue(runA, nv[0].(VarRef)), tk.ToInt(p.GetVarValue(runA, nv[1].(VarRef)), 0), tk.ToInt(p.GetVarValue(runA, nv[2].(VarRef)), 0))
	}

	if idxT == -12 { // unref
		rs, errT := tk.GetRefValue(p.GetVarValueGlobal(runA, vA.Value.(VarRef)))

		if errT != nil {
			return tk.Undefined
		}

		return rs
	}

	if idxT == -15 { // ref
		return tk.Undefined
	}

	if idxT == 3 { // normal variables
		nv, ok := p.RootFunc.Vars[vA.Value.(string)]

		if ok {
			return nv
		}

		return tk.Undefined

	}

	return tk.Undefined

}

func (p *XieVM) GetVar(runA *RunningContext, keyA string) interface{} {
	if runA == nil {
		runA = p.Running
	}

	lenT := runA.FuncStack.Size()

	for idxT := lenT - 1; idxT >= 0; idxT-- {
		loopFunc := runA.FuncStack.PeekLayer(idxT).(*FuncContext)
		nv, ok := loopFunc.Vars[keyA]

		if ok {
			return nv
		}
	}

	nv, ok := p.RootFunc.Vars[keyA]

	if ok {
		return nv
	}

	return tk.Undefined

}

// vA.Ref < -99 to get current function layer
func (p *XieVM) GetVarLayer(runA *RunningContext, vA VarRef) int {
	if runA == nil {
		runA = p.Running
	}

	idxT := vA.Ref

	if idxT < -99 {
		lenT := runA.FuncStack.Size()
		return lenT
	}

	if idxT < 0 {
		return idxT
	}

	if idxT == 3 { // normal variables
		lenT := runA.FuncStack.Size()

		for idxT := lenT - 1; idxT >= 0; idxT-- {
			loopFunc := runA.FuncStack.PeekLayer(idxT).(*FuncContext)
			_, ok := loopFunc.Vars[vA.Value.(string)]

			if ok {
				return idxT
			}
		}

		_, ok := p.RootFunc.Vars[vA.Value.(string)]

		if ok {
			return 0
		}

		return -1

	}

	return -999

}

func (p *XieVM) SetVar(runA *RunningContext, refA interface{}, setValueA interface{}) error {
	if runA == nil {
		runA = p.Running
	}

	// tk.Pln(refA, "->", setValueA)
	if refA == nil {
		return fmt.Errorf("nil parameter")
	}

	var refT VarRef

	c1, ok := refA.(int)

	if ok {
		refT = VarRef{Ref: c1, Value: nil}
	} else {
		s1, ok := refA.(string)

		if ok {
			refT = VarRef{Ref: 3, Value: s1}
		} else {
			r1, ok := refA.(VarRef)

			if ok {
				refT = r1
			} else {

				r2, ok := refA.(*VarRef)

				if ok {
					refT = *r2
				}
			}
		}
	}

	refIntT := refT.Ref

	currentFunc := p.GetCurrentFuncContext(runA)

	if refIntT == -2 { // $drop
		return nil
	}

	if refIntT == -4 { // $pln
		fmt.Println(setValueA)
		return nil
	}

	if refIntT == -5 { // $tmp
		currentFunc.Tmp = setValueA
		return nil
	}

	if refIntT == -6 { // $push
		p.Stack.Push(setValueA)
		return nil
	}

	if refIntT == -11 { // $seq
		GlobalsG.SyncSeq.Reset(tk.ToInt(setValueA, 0))
		return nil
	}

	if refIntT == -17 { // regs
		p.Regs[refT.Value.(int)] = setValueA
		return nil
	}

	if refIntT == -12 { // unref
		return nil
	}

	if refIntT == -15 { // ref
		errT := tk.SetByRef(p.GetVarValue(runA, refT.Value.(VarRef)), setValueA)

		if errT != nil {
			return errT
		}

		return nil
	}

	if refIntT != 3 {
		return fmt.Errorf("unsupported var reference")
	}

	lenT := runA.FuncStack.Size()

	keyT := refT.Value.(string)

	for idxT := lenT - 1; idxT >= 0; idxT-- {
		loopFunc := runA.FuncStack.PeekLayer(idxT).(*FuncContext)
		_, ok := loopFunc.Vars[keyT]

		if ok {
			loopFunc.Vars[keyT] = setValueA

			return nil
		}
	}

	_, ok = p.RootFunc.Vars[keyT]

	if ok {
		p.RootFunc.Vars[keyT] = setValueA
		return nil
	}

	currentFunc.Vars[keyT] = setValueA

	return nil
}

func (p *XieVM) SetVarLocal(runA *RunningContext, refA interface{}, setValueA interface{}) error {
	if runA == nil {
		runA = p.Running
	}

	// tk.Pln(refA, "->", setValueA)
	if refA == nil {
		return fmt.Errorf("nil parameter")
	}

	var refT VarRef

	c1, ok := refA.(int)

	if ok {
		refT = VarRef{Ref: c1, Value: nil}
	} else {
		s1, ok := refA.(string)

		if ok {
			refT = VarRef{Ref: 3, Value: s1}
		} else {
			r1, ok := refA.(VarRef)

			if ok {
				refT = r1
			} else {

				r2, ok := refA.(*VarRef)

				if ok {
					refT = *r2
				}
			}
		}
	}

	refIntT := refT.Ref

	currentFunc := p.GetCurrentFuncContext(runA)

	if refIntT == -2 { // $drop
		return nil
	}

	if refIntT == -4 { // $pln
		fmt.Println(setValueA)
		return nil
	}

	if refIntT == -5 { // $tmp
		currentFunc.Tmp = setValueA
		return nil
	}

	if refIntT == -6 { // $push
		p.Stack.Push(setValueA)
		return nil
	}

	if refIntT == -11 { // $seq
		GlobalsG.SyncSeq.Reset(tk.ToInt(setValueA, 0))
		return nil
	}

	if refIntT == -17 { // regs
		p.Regs[refT.Value.(int)] = setValueA
		return nil
	}

	if refIntT == -12 { // unref
		return nil
	}

	if refIntT == -15 { // ref
		errT := tk.SetByRef(p.GetVarValue(runA, refT.Value.(VarRef)), setValueA)

		if errT != nil {
			return errT
		}

		return nil
	}

	if refIntT != 3 {
		return fmt.Errorf("unsupported var reference")
	}

	keyT := refT.Value.(string)

	currentFunc.Vars[keyT] = setValueA

	return nil
}

func (p *XieVM) SetVarGlobal(refA interface{}, setValueA interface{}) error {
	// tk.Plv(refA)
	if refA == nil {
		return fmt.Errorf("nil parameter")
	}

	var refT VarRef

	c1, ok := refA.(int)

	if ok {
		refT = VarRef{Ref: c1, Value: nil}
	} else {
		s1, ok := refA.(string)

		if ok {
			refT = VarRef{Ref: 3, Value: s1}
		} else {
			r1, ok := refA.(VarRef)

			if ok {
				refT = r1
			} else {

				r2, ok := refA.(*VarRef)

				if ok {
					refT = *r2
				}
			}
		}
	}

	refIntT := refT.Ref

	currentFunc := p.GetFuncContext(nil, 0)

	if refIntT == -2 { // $drop
		return nil
	}

	if refIntT == -4 { // $pln
		fmt.Println(setValueA)
		return nil
	}

	if refIntT == -5 { // $tmp
		currentFunc.Tmp = setValueA
		return nil
	}

	if refIntT == -6 { // $push
		p.Stack.Push(setValueA)
		return nil
	}

	if refIntT == -11 { // $seq
		GlobalsG.SyncSeq.Reset(tk.ToInt(setValueA, 0))
		return nil
	}

	if refIntT == -17 { // regs
		p.Regs[refT.Value.(int)] = setValueA
		return nil
	}

	if refIntT == -12 { // unref
		return nil
	}

	if refIntT == -15 { // ref
		return nil
	}

	if refIntT != 3 {
		return fmt.Errorf("unsupported var reference")
	}

	keyT := refT.Value.(string)

	currentFunc.Vars[keyT] = setValueA

	return nil
}

// func (p *XieVM) SetVarQuick(keyA string, vA interface{}) error {

// 	lenT := p.FuncStack.Size()

// 	for idxT := lenT - 1; idxT >= 0; idxT-- {
// 		currentFunc := p.FuncStack.PeekLayer(idxT).(*FuncContext)

// 		_, ok := currentFunc.Vars[keyA]

// 		if ok {
// 			currentFunc.Vars[keyA] = vA

// 			return nil
// 		}
// 	}

// 	currentFunc := p.FuncStack.Peek().(*FuncContext)

// 	currentFunc.Vars[keyA] = vA

// 	return nil
// }

// func IsNoResult(inputA interface{}) bool {
// 	if inputA == nil {
// 		return false
// 	}

// 	nv, ok := inputA.(error)
// 	if !ok {
// 		return false
// 	}

// 	if nv.Error() == "no result" {
// 		return true
// 	}

// 	return false
// }

func (p *XieVM) Load(runA *RunningContext, codeA interface{}) error {
	if runA == nil {
		runA = p.Running
	}

	return runA.Load(codeA)
}

func (p *XieVM) LoadCompiled(runA *RunningContext, compiledA *CompiledCode) interface{} {
	if runA == nil {
		runA = p.Running
	}

	return runA.LoadCompiled(compiledA)
	// tk.Plv(compiledA)

	// // load labels
	// originalLenT := len(p.CodeList)

	// for k, _ := range compiledA.Labels {
	// 	_, ok := p.runningContext.Labels[k]

	// 	if ok {
	// 		return fmt.Errorf("duplicate label: %v", k)
	// 	}
	// }

	// for k, v := range compiledA.Labels {
	// 	p.runningContext.Labels[k] = originalLenT + v
	// }

	// // load codeList, instrList
	// p.Source = append(p.Source, compiledA.Source...)

	// p.CodeList = append(p.CodeList, compiledA.CodeList...)
	// // for _, v := range compiledA.CodeList {
	// // 	p.CodeList = append(p.CodeList, v)
	// // }

	// p.InstrList = append(p.InstrList, compiledA.InstrList...)
	// // for _, v := range compiledA.InstrList {
	// // 	p.InstrList = append(p.InstrList, v)
	// // }

	// for k, v := range compiledA.CodeSourceMap {
	// 	p.CodeSourceMap[originalLenT+k] = originalLenT + v
	// }

	// return nil
}

func (p *XieVM) Errf(runA *RunningContext, formatA string, argsA ...interface{}) error {
	if runA == nil {
		runA = p.Running
	}

	// tk.Pl("dbg: %v", tk.ToJSONX(p, "-sort"))
	// if p.VerbosePlusM {
	// 	tk.Pl(fmt.Sprintf("TXERROR:(Line %v: %v) ", p.CodeSourceMapM[p.CodePointerM]+1, tk.LimitString(p.SourceM[p.CodeSourceMapM[p.CodePointerM]], 50))+formatA, argsA...)
	// }

	return runA.Errf(formatA, argsA...)

	// return fmt.Errorf(fmt.Sprintf("(Line %v: %v) ", p.Running.CodeSourceMap[p.Running.CodePointer]+1, tk.LimitString(p.Running.Source[p.Running.CodeSourceMap[p.Running.CodePointer]], 50))+formatA, argsA...)
}

func (p *RunningContext) Errf(formatA string, argsA ...interface{}) error {
	return fmt.Errorf(fmt.Sprintf("(Line %v: %v) ", p.CodeSourceMap[p.CodePointer]+1, tk.LimitString(p.Source[p.CodeSourceMap[p.CodePointer]], 50))+formatA, argsA...)
}

func (p *XieVM) ParamsToStrs(runA *RunningContext, v *Instr, fromA int) []string {
	if runA == nil {
		runA = p.Running
	}

	lenT := len(v.Params)

	sl := make([]string, 0, lenT)

	for i := fromA; i < lenT; i++ {
		sl = append(sl, tk.ToStr(p.GetVarValue(runA, v.Params[i])))
	}

	return sl
}

func (p *XieVM) ParamsToList(runA *RunningContext, v *Instr, fromA int) []interface{} {
	if runA == nil {
		runA = p.Running
	}

	lenT := len(v.Params)

	sl := make([]interface{}, 0, lenT)

	for i := fromA; i < lenT; i++ {
		sl = append(sl, p.GetVarValue(runA, v.Params[i]))
	}

	return sl
}

type CallStruct struct {
	Type          int // 1: fastCall, 2: (normal) call
	ReturnPointer int
	ReturnRef     VarRef
	Value         interface{}
}

type RangeStruct struct {
	Iterator   tk.Iterator
	LoopIndex  int
	BreakIndex int
}

type LoopStruct struct {
	Cond       interface{}
	LoopIndex  int
	BreakIndex int
	LoopInstr  *Instr
}

func EvalCondition(condA interface{}, vmA *XieVM, runA *RunningContext) interface{} {
	var resultT, ok bool
	switch nv := condA.(type) {
	case bool:
		resultT = nv
	case string:
		rs := QuickEval(nv, vmA, runA)

		resultT, ok = rs.(bool)

		if !ok {
			return fmt.Errorf("unsupport condition type: %T(%#v)", condA, condA)
		}
	case VarRef:
		typeT := nv.Ref

		if typeT == -10 {
			rs := QuickEval(tk.ToStr(nv.Value), vmA, runA)

			resultT, ok = rs.(bool)

			if !ok {
				return fmt.Errorf("unsupport condition type: %T(%#v)", condA, condA)
			}
		} else if typeT == -3 {
			nv1, ok := nv.Value.(bool)

			if ok {
				resultT = nv1
			} else {
				rs := QuickEval(tk.ToStr(nv.Value), vmA, runA)

				resultT, ok = rs.(bool)

				if !ok {
					return fmt.Errorf("unsupport condition type: %T(%#v)", condA, condA)
				}
			}
		} else {
			rs := vmA.GetVarValue(runA, nv)

			resultT, ok = rs.(bool)

			if !ok {
				return fmt.Errorf("unsupport condition type: %T(%#v)", condA, condA)
			}
		}
	case *VarRef:
		typeT := nv.Ref

		if typeT == -10 {
			rs := QuickEval(tk.ToStr(nv.Value), vmA, runA)

			resultT, ok = rs.(bool)

			if !ok {
				return fmt.Errorf("unsupport condition type: %T(%#v)", condA, condA)
			}
		} else if typeT == -3 {
			nv1, ok := nv.Value.(bool)

			if ok {
				resultT = nv1
			} else {
				rs := QuickEval(tk.ToStr(nv.Value), vmA, runA)

				resultT, ok = rs.(bool)

				if !ok {
					return fmt.Errorf("unsupport condition type: %T(%#v)", condA, condA)
				}
			}
		} else {
			rs := vmA.GetVarValue(runA, *nv)

			resultT, ok = rs.(bool)

			if !ok {
				return fmt.Errorf("unsupport condition type: %T(%#v)", condA, condA)
			}
		}
	default:
		return fmt.Errorf("unsupport condition type: %T(%#v)", condA, condA)
	}

	return resultT
}

// if contA == true, if p.Cond == true, return p.LoopIndex; if contA == false, if p.Cond == true, return p.BreakIndex; ...
func (p *LoopStruct) ContinueCheck(contA bool, vmA *XieVM, runA *RunningContext) interface{} {
	var resultT = EvalCondition(p.Cond, vmA, runA)

	if tk.IsError(resultT) {
		return fmt.Errorf("unsupport condition type: %T(%#v)", p.Cond, p.Cond)
	}

	resultBoolT := resultT.(bool)

	if contA {
		if resultBoolT {
			return p.LoopIndex
		} else {
			return p.BreakIndex
		}
	}

	if !resultBoolT {
		return p.LoopIndex
	} else {
		return p.BreakIndex
	}
}

func NewObject(p *XieVM, r *RunningContext, typeA string, argsA ...interface{}) interface{} {
	argsT := make([]interface{}, 0, len(argsA))

	makeT := false

	for _, v := range argsA {
		// v = p.GetVarValue(r, v)

		nv, ok := v.(string)

		if ok {
			if nv == "-make" {
				makeT = true
				continue
			}
		}

		argsT = append(argsT, v)
	}

	var rs interface{}

	typeT := strings.ToLower(typeA)

	switch typeT {
	case "tk":
		if makeT {
			rs = tk.TK{Version: tk.VersionG}
		} else {
			rs = tk.NewTK()
		}
	case "postdata", "url.values":
		if makeT {
			rs = url.Values{}
		} else {
			rs = &url.Values{}
		}
	case "bool":
		if makeT {
			rs = false
		} else {
			rs = new(bool)
		}
	case "int":
		if makeT {
			rs = 0
		} else {
			rs = new(int)
		}
	case "int64":
		if makeT {
			rs = int64(0)
		} else {
			rs = new(int64)
		}
	case "uint64":
		if makeT {
			rs = uint64(0)
		} else {
			rs = new(uint64)
		}
	case "byte":
		if makeT {
			rs = byte(0)
		} else {
			rs = new(byte)
		}
	case "rune":
		if makeT {
			rs = rune(0)
		} else {
			rs = new(rune)
		}
	case "float", "float64":
		if makeT {
			rs = float64(0.0)
		} else {
			rs = new(float64)
		}
	case "float32":
		if makeT {
			rs = float32(0.0)
		} else {
			rs = new(float32)
		}
	case "str", "string":
		if makeT {
			rs = ""
		} else {
			rs = new(string)
		}
	case "bytelist": // 后面可接多个字节，其中可以有字节数组或字符串（会逐一加入字节列表中），-make参数不会加入
		blT := make([]byte, 0)

		for _, vvv := range argsT {
			nv, ok := vvv.([]byte)

			if ok {
				for _, vvvj := range nv {
					blT = append(blT, vvvj)
				}
			} else {
				nsv, ok := vvv.(string)
				if ok {
					blT = append(blT, []byte(nsv)...)
				} else {
					blT = append(blT, tk.ToByte(vvv, 0))
				}
			}
		}

		if makeT {
			rs = blT
		} else {
			rs = &blT
		}
	case "bytesbuffer", "bytesBuf":
		if makeT {
			rs = bytes.Buffer{}
		} else {
			rs = new(bytes.Buffer)
		}
	case "stringbuffer", "strbuf", "strings.builder":
		if makeT {
			rs = strings.Builder{}

			if len(argsT) > 0 {
				// (&(rs.(strings.Builder))).(*strings.Builder).WriteString(tk.ToStr(argsT[0]))
			}
		} else {
			rs = new(strings.Builder)
			if len(argsT) > 0 {
				rs.(*strings.Builder).WriteString(tk.ToStr(argsT[0]))
			}
		}

	case "reader": // 读取参数中的类型，自动判断后统一转为Go语言中的io.Reader
		if len(argsT) < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		vs1 := argsT[0]

		switch nv := vs1.(type) {
		case string:
			rs = strings.NewReader(nv)
		case []byte:
			rs = bytes.NewReader(nv)
		case *os.File:
			rs = nv
		default:
			return p.Errf(r, "type not supported: %T(%v)", vs1, vs1)
		}

	case "filereader": // 打开字符串参数指定的路径名的文件，转为io.Reader/FILE
		if len(argsT) < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		vs1 := tk.ToStr(argsT[0])

		fileT, errT := os.Open(vs1)

		if errT != nil {
			rs = errT
		} else {
			rs = fileT
		}

	case "bufio.reader": // 打开支持io.Reader接口的对象为缓冲io对象
		if len(argsT) < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		vs1 := argsT[0]

		nv, ok := vs1.(io.Reader)

		if !ok {
			rs = fmt.Errorf("invalid type %T(%v)", vs1, vs1)
		} else {
			rs = bufio.NewReader(nv)
		}

	case "time":
		if makeT {
			rs = time.Now()
		} else {
			rs = new(time.Time)
		}
	case "mutex", "lock": // 同步锁
		if makeT {
			rs = sync.RWMutex{}
		} else {
			rs = new(sync.RWMutex)
		}
	case "waitgroup": // 同步等待组
		// var wg sync.WaitGroup
		// wg.Add(1)
		// go func() {
		// 	defer wg.Done()
		// 	...
		// }()
		// wg.Wait()
		if makeT {
			rs = sync.WaitGroup{}
		} else {
			rs = new(sync.WaitGroup)
		}
	case "mux": // http请求处理路由器
		if makeT {
			rs = http.ServeMux{}
		} else {
			rs = http.NewServeMux()
		}
	case "seq": // 序列生成器（自动增长的整数序列，一般用于需要唯一性ID时）
		if makeT {
			rs = tk.Seq{}
		} else {
			rs = tk.NewSeq()
		}
	case "instr": // 指令
		if len(argsT) < 1 {
			if makeT {
				rs = Instr{Code: 12, Params: []VarRef{}} // invalid instr
			} else {
				rs = &Instr{Code: 12, Params: []VarRef{}}
			}
		} else {
			vs1 := tk.ToStr(argsT[0])
			rs1 := Compile(vs1)

			if tk.IsError(rs1) {
				return p.Errf(r, "failed to compile the instr code(编译指令代码失败): %v", rs1)
			}

			rs1n := rs1.(*CompiledCode)

			if len(rs1n.InstrList) < 1 {
				return p.Errf(r, "failed to compile the instr code(编译指令代码失败): %v", "empty code(没有有效代码)")
			}

			if makeT {
				rs = rs1n.InstrList[0]
			} else {
				rs = &rs1n.InstrList[0]
			}
		}

	case "messagequeue", "syncqueue": // 线程安全的先进先出队列
		if makeT {
			rs = tk.SyncQueue{}
		} else {
			rs = tk.NewSyncQueue()
		}
	case "mailsender": // 邮件发送客户端
		hostT := strings.TrimSpace(p.GetSwitchVarValueI(r, argsT, "-host=", ""))
		portT := strings.TrimSpace(p.GetSwitchVarValueI(r, argsT, "-port=", "25"))
		userT := strings.TrimSpace(p.GetSwitchVarValueI(r, argsT, "-user=", ""))
		passT := strings.TrimSpace(p.GetSwitchVarValueI(r, argsT, "-pass=", ""))
		if strings.HasPrefix(passT, "740404") {
			passT = tk.DecryptStringByTXDEF(passT)
		}

		if hostT == "" {
			return p.Errf(r, "服务器地址不能为空（empty host）")
		}

		if portT == "" {
			return p.Errf(r, "端口不能为空（empty port")
		}

		if userT == "" {
			return p.Errf(r, "用户名不能为空（empty user name")
		}

		if passT == "" {
			return p.Errf(r, "口令不能为空（empty password")
		}

		rs = mailyak.New(hostT+":"+portT, tk.GetLoginAuth(userT, passT))

	case "ssh":
		hostT := strings.TrimSpace(p.GetSwitchVarValueI(r, argsT, "-host=", ""))
		portT := strings.TrimSpace(p.GetSwitchVarValueI(r, argsT, "-port=", "25"))
		userT := strings.TrimSpace(p.GetSwitchVarValueI(r, argsT, "-user=", ""))
		passT := strings.TrimSpace(p.GetSwitchVarValueI(r, argsT, "-pass=", ""))
		if strings.HasPrefix(passT, "740404") {
			passT = tk.DecryptStringByTXDEF(passT)
		}
		// v1p = 0
		// v2 := tk.ToStr(p.GetVarValue(argsT[v1p]))
		// v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
		// v4 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))
		// v5 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+3]))
		// if strings.HasPrefix(v5, "740404") {
		// 	v5 = tk.DecryptStringByTXDEF(v5)
		// }

		sshT, errT := tk.NewSSHClient(hostT, portT, userT, passT)

		if errT != nil {
			return p.Errf(r, "failed to create ssh object: %v", errT)
		}

		rs = sshT
	case "gui":
		rs = p.GetVar(r, "guiG")
	// case "quickStringDelegate": // quickStringDelegate中，CodePointerM并不跳转（除非有移动其的指令执行）
	// 	if instrT.ParamLen < 3 {
	// 		return p.Errf(r, "not enough paramters")
	// 	}

	// 	v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))

	// 	var deleT tk.QuickDelegate

	// 	// same as fastCall
	// 	deleT = func(strA string) string {
	// 		pointerT := p.CodePointerM

	// 		p.Push(strA)

	// 		tmpPointerT := v2
	// 		p.CodePointerM = tmpPointerT
	// 		for {
	// 			rs := p.RunLine(tmpPointerT)

	// 			nv, ok := rs.(int)

	// 			if ok {
	// 				tmpPointerT = nv
	// 				p.CodePointerM = tmpPointerT
	// 				continue
	// 			}

	// 			nsv, ok := rs.(string)

	// 			if ok {
	// 				if tk.IsErrStr(nsv) {
	// 					// tmpRs := p.Pop()

	// 					p.CodePointerM = pointerT
	// 					return nsv
	// 				}

	// 				if nsv == "exit" { // 不应发生
	// 					tmpRs := p.Pop()
	// 					p.CodePointerM = pointerT
	// 					return tk.ToStr(tmpRs)
	// 				} else if nsv == "fr" {
	// 					break
	// 				}
	// 			}

	// 			tmpPointerT++
	// 			p.CodePointerM = tmpPointerT
	// 		}

	// 		// return pointerT + 1

	// 		tmpRs := p.Pop()
	// 		p.CodePointerM = pointerT
	// 		return tk.ToStr(tmpRs)
	// 	}

	// 	p.SetVar(r, pr, deleT)
	case "quickdelegate": // quickDelegate中，CodePointerM并不跳转（除非有移动其的指令执行）
		if len(argsT) < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var deleT tk.QuickVarDelegate

		codeT := argsT[0]

		if s1, ok := codeT.(string); ok {
			// s1 = strings.ReplaceAll(s1, "~~~", "`")
			compiledT := Compile(s1)

			if tk.IsError(compiledT) {
				return p.Errf(r, "failed to compile the quick delegate code: %v", compiledT)
			}

			codeT = compiledT
		}

		cp1, ok := codeT.(*CompiledCode)

		if !ok {
			return p.Errf(r, "invalid compiled object: %v", codeT)
		}

		deleT = func(argsA ...interface{}) interface{} {
			rs := RunCodePiece(p, nil, cp1, argsA, true)

			return rs
		}

		rs = deleT

	case "delegate": // delegate中，类似callFunc，将使用单独的虚拟机执行代码
		if len(argsT) < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v2 := argsT[0]

		// if nv, ok := v2.(string); ok {
		// 	v2 = strings.ReplaceAll(nv, "~~~", "`")
		// }

		vmT := NewVMQuick()

		lrs := vmT.Load(nil, v2)

		if tk.IsError(lrs) {
			return p.Errf(r, "failed to create VM: %v", lrs)
		}

		var deleT tk.QuickVarDelegate

		// like sealFunc
		deleT = func(argsA ...interface{}) interface{} {
			vmT.SetVar(nil, "inputG", argsA)

			rs := vmT.Run()

			return rs
		}

		rs = deleT
	case "image.point", "point":
		// var p1 image.Point
		if makeT {
			rs = image.Point{X: 0, Y: 0}
		} else {
			rs = new(image.Point)
		}

	case "zip":
		// vs := p.ParamsToList(r, instrT, v1p+1)
		// tk.Plv(vs)
		if makeT {
			// rs = image.Point{X:0, Y:0}
		} else {
			rs = archiver.NewZip()
		}

	default:
		if len(argsT) > 0 {
			rs = tk.NewObject(append([]interface{}{typeT}, argsT...)...)
		} else {
			rs = tk.NewObject(typeT)
		}

		// if tk.IsErrX(rs) {

		// }
		// return p.Errf(r, "unsupported object: %v", typeA)
	}

	return rs
}

func NewVar(p *XieVM, r *RunningContext, typeA string, argsA ...interface{}) interface{} {
	// argsT := make([]interface{}, 0, len(argsA))

	// makeT := false

	// for _, v := range argsA {
	// 	// v = p.GetVarValue(r, v)

	// 	nv, ok := v.(string)

	// 	if ok {
	// 		if nv == "-make" {
	// 			makeT = true
	// 			continue
	// 		}
	// 	}

	// 	argsT = append(argsT, v)
	// }

	argsLenT := len(argsA)

	var rs interface{}

	switch typeA {
	case "postData", "url.Values":
		rs = url.Values{}
	case "bool":
		if argsLenT < 1 {
			rs = false
		} else {
			rs = tk.ToBool(argsA[0])
		}
	case "int":
		if argsLenT < 1 {
			rs = 0
		} else {
			rs = tk.ToInt(argsA[0], 0)
		}
	case "int64":
		if argsLenT < 1 {
			rs = int64(0)
		} else {
			rs = int64(tk.ToInt(argsA[0], 0))
		}
	case "uint64":
		if argsLenT < 1 {
			rs = uint64(0)
		} else {
			rs = uint64(tk.ToInt(argsA[0], 0))
		}
	case "byte":
		if argsLenT < 1 {
			rs = byte(0)
		} else {
			rs = tk.ToByte(argsA[0], 0)
		}
	case "rune":
		if argsLenT < 1 {
			rs = rune(0)
		} else {
			rs = tk.ToRune(argsA[0], 0)
		}
	case "float", "float64":
		if argsLenT < 1 {
			rs = float64(0.0)
		} else {
			rs = tk.ToFloat(argsA[0], 0)
		}
	case "float32":
		if argsLenT < 1 {
			rs = float32(0.0)
		} else {
			rs = float32(tk.ToFloat(argsA[0], 0))
		}
	case "str", "string":
		if argsLenT < 1 {
			rs = ""
		} else {
			rs = tk.ToStr(argsA[0])
		}
	case "list", "array", "[]":
		blT := make([]interface{}, 0, argsLenT)

		for _, vvv := range argsA {
			nv, ok := vvv.([]interface{})

			if ok {
				for _, vvvj := range nv {
					blT = append(blT, vvvj)
				}
			} else {
				blT = append(blT, vvv)
			}
		}

		rs = blT

	case "strList", "[]string": // 后面可接多个字符串，其中可以有字节数组或字符串（会逐一加入字节列表中）
		blT := make([]string, 0, argsLenT)

		for _, vvv := range argsA {
			nv, ok := vvv.([]string)

			if ok {
				for _, vvvj := range nv {
					blT = append(blT, vvvj)
				}
			} else {
				nsv, ok := vvv.(string)
				if ok {
					blT = append(blT, nsv)
				} else {
					blT = append(blT, tk.ToStr(vvv))
				}
			}
		}

		rs = blT
	case "byteList": // 后面可接多个字节，其中可以有字节数组或字符串（会逐一加入字节列表中）
		blT := make([]byte, 0)

		for _, vvv := range argsA {
			nv, ok := vvv.([]byte)

			if ok {
				for _, vvvj := range nv {
					blT = append(blT, vvvj)
				}
			} else {
				nsv, ok := vvv.(string)
				if ok {
					blT = append(blT, []byte(nsv)...)
				} else {
					nbv, ok := vvv.(byte)
					if ok {
						blT = append(blT, nbv)
					} else {
						blT = append(blT, tk.ToByte(vvv, 0))
					}
				}
			}
		}

		rs = blT
	case "runeList": // runeList
		blT := make([]rune, 0, argsLenT)

		for _, vvv := range argsA {
			nv, ok := vvv.([]rune)

			if ok {
				for _, vvvj := range nv {
					blT = append(blT, vvvj)
				}
			} else {
				nsv, ok := vvv.(rune)
				if ok {
					blT = append(blT, nsv)
				} else {
					blT = append(blT, tk.ToRune(vvv, 0))
				}
			}
		}

		rs = blT
	case "map", "{}", "map[string]interface{}":
		rs = map[string]interface{}{}
	case "strMap", "map[string]string":
		rs = map[string]string{}
	case "bytesBuffer", "bytesBuf":
		rs = new(bytes.Buffer)
	case "stringBuffer", "strBuf", "strings.Builder":
		rs = new(strings.Builder)
		if argsLenT > 0 {
			rs.(*strings.Builder).WriteString(tk.ToStr(argsA[0]))
		}
	case "time":
		if argsLenT > 0 {
			rs = tk.ToTime(argsA[0])
		} else {
			rs = time.Now()
		}
	case "mutex", "lock": // 同步锁
		rs = sync.RWMutex{}
	case "waitGroup": // 同步等待组
		// var wg sync.WaitGroup
		// wg.Add(1)
		// go func() {
		// 	defer wg.Done()
		// 	...
		// }()
		// wg.Wait()
		rs = sync.WaitGroup{}
	case "mux": // http请求处理路由器
		rs = http.ServeMux{}
	case "seq": // 序列生成器（自动增长的整数序列，一般用于需要唯一性ID时）
		rs = tk.Seq{}
	case "messageQueue", "syncQueue": // 线程安全的先进先出队列
		rs = tk.SyncQueue{}
	case "gui":
		rs = p.GetVar(r, "guiG")
	case "quickDelegate": // quickDelegate中，CodePointerM并不跳转（除非有移动其的指令执行）
		if len(argsA) < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var deleT tk.QuickVarDelegate

		codeT := argsA[0]

		if s1, ok := codeT.(string); ok {
			// s1 = strings.ReplaceAll(s1, "~~~", "`")
			compiledT := Compile(s1)

			if tk.IsError(compiledT) {
				return p.Errf(r, "failed to compile the quick delegate code: %v", compiledT)
			}

			codeT = compiledT
		}

		cp1, ok := codeT.(*CompiledCode)

		if !ok {
			return p.Errf(r, "invalid compiled object: %v", codeT)
		}

		deleT = func(argsA ...interface{}) interface{} {
			rs := RunCodePiece(p, nil, cp1, argsA, true)

			return rs
		}

		rs = deleT

	case "delegate": // delegate中，类似callFunc，将使用单独的虚拟机执行代码
		if len(argsA) < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v2 := argsA[0]

		// if nv, ok := v2.(string); ok {
		// 	v2 = strings.ReplaceAll(nv, "~~~", "`")
		// }

		vmT := NewVMQuick()

		lrs := vmT.Load(nil, v2)

		if tk.IsError(lrs) {
			return p.Errf(r, "failed to create VM: %v", lrs)
		}

		var deleT tk.QuickVarDelegate

		// like sealFunc
		deleT = func(argsA ...interface{}) interface{} {
			vmT.SetVar(nil, "inputG", argsA)

			rs := vmT.Run()

			return rs
		}

		rs = deleT
	case "image.Point", "point":
		// var p1 image.Point
		rs = image.Point{X: 0, Y: 0}

	case "zip":
		rs = archiver.NewZip()

	default:
		return p.Errf(r, "unsupported object: %v", typeA)
	}

	return rs
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

func GoCall(codeA interface{}, inputA ...interface{}) error {
	vmObjT := NewVM(codeA)

	if tk.IsError(vmObjT) {
		return vmObjT.(error)
	}

	vmT := vmObjT.(*XieVM)

	if len(inputA) > 0 {
		vmT.SetVar(nil, "inputG", inputA)
	}

	// var lrs error

	// nvr, ok := codeA.(*RunningContext)

	// if ok {
	// 	lrs = vmT.Load(nvr, nil)
	// } else {
	// 	lrs = vmT.Load(nil, codeA)
	// }

	// tk.Plv("lrs:", lrs)

	// if tk.IsError(lrs) {
	// 	return lrs
	// }

	go vmT.Run()

	return nil
}

// func (p *XieVM) GetVarRef(runA *RunningContext, vA VarRef) interface{} {
// 	idxT := vA.Ref

// 	if idxT == -2 {
// 		return tk.Undefined
// 	}

// 	if idxT == -3 {
// 		return tk.Undefined
// 	}

// 	if idxT == -8 {
// 		return tk.Undefined
// 	}

// 	if idxT == -7 {
// 		return tk.Undefined
// 	}

// 	if idxT == -5 {
// 		return &(p.GetCurrentFuncContext(runA).Tmp)
// 	}

// 	if idxT == -11 { // 自增长序号无法寻址
// 		return tk.Undefined
// 	}

// 	if idxT == -6 {
// 		return tk.Undefined
// 	}

// 	if idxT == -9 {
// 		return tk.Undefined
// 	}

// 	if idxT == -10 {
// 		return tk.Undefined
// 	}

// 	if idxT < 0 {
// 		return tk.Undefined
// 	}

// 	if idxT == 3 { // normal variables
// 		lenT := runA.FuncStack.Size()

// 		for idxT := lenT - 1; idxT >= 0; idxT-- {
// 			loopFunc := runA.FuncStack.PeekLayer(idxT).(*FuncContext)
// 			nv, ok := loopFunc.Vars[vA.Value.(string)]

// 			if ok {
// 				return &loopFunc.Vars[vA.Value.(string)]
// 			}
// 		}

// 		nv, ok := p.RootFunc.Vars[vA.Value.(string)]

// 		if ok {
// 			return nv
// 		}

// 		return tk.Undefined

// 	}

// 	return tk.Undefined
// }

func QuickRunPiece(vmA *XieVM, runA *RunningContext, codeA interface{}) (rst interface{}) {
	if runA == nil {
		runA = NewRunningContext().(*RunningContext)
	}

	nc, ok := codeA.(*CompiledCode)

	if !ok {
		nip, ok := codeA.(*Instr)

		if ok {
			nc = &CompiledCode{InstrList: []Instr{*nip}}
		} else {
			ni, ok := codeA.(Instr)

			if ok {
				nc = &CompiledCode{InstrList: []Instr{ni}}
			} else {
				nnc := Compile(tk.ToStr(codeA))

				if tk.IsError(nnc) {
					return fmt.Errorf("failed to compile code: %v", nnc)
				}

				nc = nnc.(*CompiledCode)
			}
		}
	}

	lenT := len(nc.InstrList)

	tmpPointerT := 0

	for (tmpPointerT >= 0) && (tmpPointerT < lenT) {
		rs := RunInstr(vmA, runA, &nc.InstrList[tmpPointerT])

		nv, ok := rs.(int)

		if ok {
			tmpPointerT = nv
			continue
		}

		if tk.IsError(rs) {
			return rs
		}

		nsv, ok := rs.(string)

		if ok {
			if nsv == "exit" {
				return tk.Undefined
			}
		}

		tmpPointerT++
	}

	return tk.Undefined
}

// 避免使用任何跳转到当前代码外的指令
// return value is not required normally, but errors
// normally return tk.Undefined
func (p *XieVM) QuickRunCode(codeA interface{}) interface{} {
	nr := NewRunningContext(codeA)

	if tk.IsError(nr) {
		return nr
	}

	nrr := nr.(*RunningContext)

	lenT := len(nrr.InstrList)

	tmpPointerT := 0

	for (tmpPointerT >= 0) && (tmpPointerT < lenT) {
		rs := RunInstr(p, p.Running, &nrr.InstrList[tmpPointerT])

		nv, ok := rs.(int)

		if ok {
			tmpPointerT = nv
			continue
		}

		if tk.IsError(rs) {
			return rs
		}

		nsv, ok := rs.(string)

		if ok {
			if nsv == "exit" {
				nrr.RunDeferUpToRoot(p)
				return tk.Undefined
			}
		}

		tmpPointerT++
	}

	return tk.Undefined
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

func leAppendFile(fileNameA string) error {
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

	leBufG = append(leBufG, tk.SplitLines(strT)...)
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

func leSort(descentA ...bool) error {
	descentT := false
	if len(descentA) > 0 {
		descentT = descentA[0]
	}

	if leBufG == nil {
		leClear()
	}

	if leBufG == nil {
		return tk.Errf("buffer not initalized")
	}

	if descentT {
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

func RunInstr(p *XieVM, r *RunningContext, instrA *Instr) (resultR interface{}) {
	defer func() {
		if r1 := recover(); r1 != nil {
			// tk.Printfln("exception: %v", r)
			if r.ErrorHandler > -1 {
				p.SetVarGlobal("lastLineG", r.CodeSourceMap[r.CodePointer]+1)
				p.SetVarGlobal("errorMessageG", tk.ToStr(r))
				p.SetVarGlobal("errorDetailG", p.Errf(r, "runtime error: %v\n%v", r, string(debug.Stack())))

				// p.Stack.Push(p.Errf(r, "runtime error: %v\n%v", r, string(debug.Stack())))
				// p.Stack.Push(tk.ToStr(r))
				// p.Stack.Push(r.CodeSourceMap[r.CodePointer] + 1)
				resultR = r.ErrorHandler
				return
			}

			resultR = fmt.Errorf("runtime exception: %v\n%v", r1, string(debug.Stack()))

			return
		}
	}()

	var instrT *Instr = instrA

	// if p.VerbosePlusM {
	// tk.Plv(instrT)
	// }

	if instrT == nil {
		return p.Errf(r, "nil instr: %v", instrT)
	}

	cmdT := instrT.Code

	switch cmdT {
	case 12: // invalidInstr
		return p.Errf(r, "invalid instr: %v", instrT.Params[0].Value)
	case 100: // version
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		p.SetVar(r, pr, VersionG)

		return ""
	case 101: // pass
		return ""

	case 102: // debug
		outT := map[string]interface{}{
			"VM":                    p,
			"CurrentRunningContext": r,
			"Instr":                 instrT,
		}

		tk.Pl("[DEBUG INFO] %v", tk.ToJSONX(outT, "-indent", "-sort"))

		return ""

	case 103: // debugInfo
		var pr any = -5
		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		outT := map[string]interface{}{
			"VM":                    p,
			"CurrentRunningContext": r,
			"Instr":                 instrT,
		}

		p.SetVar(r, pr, tk.ToJSONX(outT, "-indent", "-sort"))

		return ""
	case 104: // varInfo
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0
		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, fmt.Sprintf("[变量信息]: %v(行：%v) -> (%T) %v", instrT.Params[v1p].Ref, p.GetVarLayer(r, instrT.Params[v1p]), v1, v1))

		return ""
	case 106: // onError
		if instrT.ParamLen < 1 {
			r.ErrorHandler = -1
			return ""
		}

		r.ErrorHandler = tk.ToInt(p.GetVarValue(r, instrT.Params[0]))

		return ""

	case 107: // dumpf
		// if instrT.ParamLen < 1 {
		tk.Dump(p, r)
		return ""
	// }

	// v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[0]))

	// if v1 == "all" {
	// 	tk.Dump(p, r)
	// } else if v1 == "vars" {
	// 	for k, v := range p.FuncContextM.VarsLocalMapM {
	// 		tk.Dumpf("%v -> %v", p.VarNameMapM[k], (*(p.FuncContextM.VarsM))[v])
	// 	}

	// } else if v1 == "labels" {
	// 	for k, v := range p.LabelsM {
	// 		tk.Dumpf("%v -> %v/%v (%v)", p.VarNameMapM[k], v, p.CodeSourceMapM[v], tk.LimitString(p.SourceM[p.CodeSourceMapM[v]], 50))
	// 	}

	// } else {
	// 	tk.Dumpf(v1, p.ParamsToList(r, instrT, 1)...)
	// }

	// return ""
	case 109: // defer
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0])

		// tk.Plo(v1)
		// return ""

		nvi, ok := v1.(Instr)

		if ok {
			p.GetCurrentFuncContext(r).DeferStack.Push(nvi)
			return ""
		}

		nvc, ok := v1.(*CompiledCode)

		if ok {
			p.GetCurrentFuncContext(r).DeferStack.Push(nvc)
			return ""
		}

		var nvs string

		if instrT.ParamLen > 1 {
			nvs = strings.TrimSpace(instrT.Line)
		} else {
			nvs = tk.ToStr(v1)
		}

		nc := Compile(nvs)

		if tk.IsError(nc) {
			return p.Errf(r, "failed to compile defer code: %v", nc)
		}

		p.GetCurrentFuncContext(r).DeferStack.Push(nc)

		// codeT, ok := InstrNameSet[nvs]

		// if !ok {
		// 	return p.Errf(r, "unknown instruction: %v", v1)
		// }

		// instrT := Instr{Code: codeT, Cmd: nvs, Params: instrT.Params[1:], ParamLen: instrT.ParamLen - 1, Line: tk.RemoveFirstSubString(strings.TrimSpace(instrT.Line), nvs)} //&([]VarRef{})}

		// p.GetCurrentFuncContext(r).DeferStack.Push(instrT)

		return ""
	case 110: // deferStack
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		p.SetVar(r, pr, tk.ToJSONX(GetDeferStack(p, r), "-indent", "-sort"))
		return ""
	case 111: // isUndef
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, v1 == tk.Undefined)

		return ""

	case 112: // isDef
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, v1 != tk.Undefined)

		return ""

	case 113: // isNil
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, tk.IsNil(v1))

		return ""

	case 121: // test
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0])
		v2 := p.GetVarValue(r, instrT.Params[1])

		// tk.Plo("--->", v2)

		var v3 string
		var v4 string

		if instrT.ParamLen > 3 {
			v3 = tk.ToStr(p.GetVarValue(r, instrT.Params[2]))
			v4 = "(" + tk.ToStr(p.GetVarValue(r, instrT.Params[3])) + ")"
		} else if instrT.ParamLen > 2 {
			v3 = tk.ToStr(p.GetVarValue(r, instrT.Params[2]))
		} else {
			v3 = tk.ToStr(GlobalsG.SyncSeq.Get())
		}

		if v1 == v2 {
			tk.Pl("test %v%v passed", v3, v4)
		} else {
			return p.Errf(r, "test %v%v failed: %#v <-> %#v\n-----\n%v\n-----\n%v", v3, v4, v1, v2, v1, v2)
		}

		return ""

	case 122: // testByStartsWith
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0])
		v2 := p.GetVarValue(r, instrT.Params[1])

		var v3 string
		var v4 string

		if instrT.ParamLen > 3 {
			v3 = tk.ToStr(p.GetVarValue(r, instrT.Params[2]))
			v4 = "(" + tk.ToStr(p.GetVarValue(r, instrT.Params[3])) + ")"
		} else if instrT.ParamLen > 2 {
			v3 = tk.ToStr(p.GetVarValue(r, instrT.Params[2]))
		} else {
			v3 = tk.ToStr(GlobalsG.SyncSeq.Get())
		}

		if strings.HasPrefix(tk.ToStr(v1), tk.ToStr(v2)) {
			tk.Pl("test %v%v passed", v3, v4)
		} else {
			return p.Errf(r, "test %v%v failed: %#v <-> %#v", v3, v4, v1, v2)
		}

		return ""

	case 123: // testByReg
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0])
		v2 := p.GetVarValue(r, instrT.Params[1])

		var v3 string
		var v4 string

		if instrT.ParamLen > 3 {
			v3 = tk.ToStr(p.GetVarValue(r, instrT.Params[2]))
			v4 = "(" + tk.ToStr(p.GetVarValue(r, instrT.Params[3])) + ")"
		} else if instrT.ParamLen > 2 {
			v3 = tk.ToStr(p.GetVarValue(r, instrT.Params[2]))
		} else {
			v3 = tk.ToStr(GlobalsG.SyncSeq.Get())
		}

		if tk.RegMatchX(tk.ToStr(v1), tk.ToStr(v2)) {
			tk.Pl("test %v%v passed", v3, v4)
		} else {
			return p.Errf(r, "test %v%v failed: %#v <-> %#v --- %v", v3, v4, v1, v2, v2)
		}

		return ""
	case 131: // typeOf

		var pr interface{} = -5

		var v1 interface{}

		if instrT.ParamLen < 1 {
			v1 = p.Stack.Peek()
		} else if instrT.ParamLen < 2 {
			v1 = p.GetVarValue(r, instrT.Params[0])
		} else {
			pr = instrT.Params[0]
			v1 = p.GetVarValue(r, instrT.Params[1])
		}

		p.SetVar(r, pr, fmt.Sprintf("%T", v1))

		return ""
	case 141: // layer
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		p.SetVar(r, pr, r.FuncStack.Size())

		return ""

	case 151: // loadCode

		var pr any = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		codeT := p.GetVarValue(r, instrT.Params[v1p])

		rsT := p.Load(r, codeT)

		p.SetVar(r, pr, rsT)

		return ""

	case 152: // loadGel

		// var pr any = -5
		// v1p := 0

		// if instrT.ParamLen > 1 {
		pr := instrT.Params[0]
		v1p := 1
		// }

		urlT := p.GetVarValue(r, instrT.Params[v1p])

		vs := p.ParamsToStrs(r, instrT, v1p+1)

		keyT := tk.GetSwitch(vs, "-key=", "")
		ifFileT := tk.IfSwitchExists(vs, "-file")

		var fcT interface{}

		if ifFileT {
			fcT = tk.LoadText(tk.ToStr(urlT))
		} else {
			fcT = tk.GetWeb(tk.ToStr(urlT))
		}

		if tk.IsErrX(fcT) {
			p.SetVar(r, pr, fmt.Errorf("failed to load gel: %v", tk.GetErrStrX(fcT)))
			return
		}

		fcsT := tk.ToStr(fcT)

		if strings.HasPrefix(fcsT, "740404") || keyT != "" {
			fcsT = tk.DecryptStringByTXDEF(fcsT, keyT)
			if tk.IsErrX(fcsT) {
				p.SetVar(r, pr, fmt.Errorf("failed to extract gel: %v", tk.GetErrStrX(fcT)))
				return
			}
		}

		rsT := Compile(fcsT)

		if tk.IsError(rsT) {
			p.SetVar(r, pr, fmt.Errorf("failed to compile gel: %v", rsT))
			return
		}

		p.SetVar(r, pr, rsT)

		return ""

	case 153: // compile
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rs := Compile(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])))

		p.SetVar(r, pr, rs)

		return ""

	case 155: // quickRun
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rs := p.QuickRunCode(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, rs)

		return ""

	case 156: // runCode
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v1p := 1

		// 可以是字符串类型的代码或编译后的代码（*CompiledCode）
		codeT := p.GetVarValue(r, instrT.Params[v1p])

		// 获取RunCode的inputA参数
		inputT := p.GetVarValue(r, GetVarRefInParams(instrT.Params, v1p+1))

		if tk.IsUndefined(inputT) {
			inputT = nil
		}

		// 获取RunCode的objA参数
		objT := p.GetVarValue(r, GetVarRefInParams(instrT.Params, v1p+2))

		obj1, ok := objT.(map[string]interface{})

		if !ok {
			obj1 = nil
		}

		// 获取RunCode的optsA参数（传入虚拟机中的argsG参数）
		// vs := p.ParamsToStrs(r, instrT, v1p+3)
		args1 := p.GetVarValue(r, GetVarRefInParams(instrT.Params, v1p+3))

		var errT error

		args2a, ok := args1.([]string)
		if !ok {
			args2i, ok := args1.([]interface{})

			if ok {
				args2a = make([]string, 0, len(args2i))
				for _, va := range args2i {
					args2a = append(args2a, tk.ToStr(va))
				}
			} else {
				args2s, ok := args1.(string)
				if ok {
					args2a, errT = tk.ParseCommandLine(args2s)

					if errT == nil {

					} else {
						args2a = []string{args2s}
					}
				} else {
					args2a = []string{tk.ToStr(args1)}
				}
			}
		}

		rs := RunCode(codeT, inputT, obj1, args2a...)

		p.SetVar(r, pr, rs)

		return ""
	case 157: // runPiece
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		codeT := p.GetVarValue(r, instrT.Params[v1p])

		rs := QuickRunPiece(p, r, codeT)

		p.SetVar(r, pr, rs)

		return ""

	case 158: // extractRun
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		rs := r.Extract(tk.ToInt(v1), tk.ToInt(v2))

		p.SetVar(r, pr, rs)

		return ""

	case 159: // extractCompiled
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		rs := r.ExtractCompiled(tk.ToInt(v1), tk.ToInt(v2))

		p.SetVar(r, pr, rs)

		return ""

	case 161: // len
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rs := tk.Len(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, rs)

		return ""
	case 170: // fatalf
		list1T := []interface{}{}

		formatT := ""

		for i, v := range instrT.Params {
			if i == 0 {
				formatT = tk.ToStr(v.Value)
				continue
			}

			list1T = append(list1T, p.GetVarValue(r, v))
		}

		fmt.Printf(formatT+"\n", list1T...)

		return "exit"

	case 180: // goto/jmp
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0])

		c1 := r.GetLabelIndex(v1)

		if c1 >= 0 {
			return c1
		}

		// c, ok := v1.(int)

		// if ok {
		// 	return c
		// }

		// s2 := tk.ToStr(v1)

		// if len(s2) > 1 && s2[0:1] == ":" {
		// 	s2a := s2[1:]
		// 	if strings.HasPrefix(s2a, "+") {
		// 		return r.CodePointer + tk.ToInt(s2a[1:])
		// 	} else if strings.HasPrefix(s2, "-") {
		// 		return r.CodePointer - tk.ToInt(s2a[1:])
		// 	} else {
		// 		labelPointerT, ok := r.Labels[s2a]

		// 		if ok {
		// 			return labelPointerT
		// 		}
		// 	}
		// }

		return p.Errf(r, "invalid label: %v", v1)

	case 191: // wait
		if instrT.ParamLen < 1 {
			return p.Errf(r, "参数不够")
		}

		v1 := p.GetVarValue(r, instrT.Params[0])

		c, ok := v1.(int)

		if ok {
			tk.Sleep(tk.ToFloat(c))
			return ""
		}

		f, ok := v1.(float64)

		if ok {
			tk.Sleep(f)
			return ""
		}

		s2, ok := v1.(string)

		if ok {
			tk.GetInputf(s2)
			return ""
		}

		wg1, ok := v1.(*sync.WaitGroup)

		if ok {
			wg1.Wait()
			return ""
		}

		ch1, ok := v1.(<-chan struct{})

		if ok {
			<-ch1
			return ""
		}

		for {
			tk.Sleep(1.0)
		}

		return p.Errf(r, "不支持的数据类型（unsupported type）：%T(%v)", v1, v1)

	case 197: // exitL
		if instrT.ParamLen < 1 {
			return "exit"
		}

		valueT := p.GetVarValue(r, instrT.Params[0])

		p.SetVar(r, "outL", valueT)

		return "exit"

	case 198: // exitfL

		list1T := []interface{}{}

		formatT := ""

		for i, v := range instrT.Params {
			if i == 0 {
				formatT = tk.ToStr(v.Value)
				continue
			}

			list1T = append(list1T, p.GetVarValue(r, v))
		}

		p.SetVar(r, "outL", fmt.Sprintf(formatT, list1T...))

		return "exit"

	case 199: // exit
		if instrT.ParamLen < 1 {
			return "exit"
		}

		valueT := p.GetVarValue(r, instrT.Params[0])

		p.SetVarGlobal("outG", valueT)

		return "exit"

	case 201: // global
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "参数不够")
		// }

		// pr := instrT.Params[0]
		// v1p := 1

		// contextT := p.CurrentFuncContextM

		// if instrT.ParamLen < 2 {
		// 	p.SetVarGlobal(pr, nil)
		// 	// contextT.VarsM[nameT] = ""
		// 	return ""
		// }

		// v1 := p.GetVarValue(r, instrT.Params[v1p])

		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough paramters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		// contextT := p.CurrentFuncContextM

		if instrT.ParamLen < 2 {
			p.SetVarGlobal(pr, nil)
			// contextT.VarsM[nameT] = ""
			return ""
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		vs := p.ParamsToList(r, instrT, v1p+1)

		var pv interface{} = NewVar(p, r, v1, vs...)

		p.SetVarGlobal(pr, pv)

		return ""

		// if v1 == "bool" {
		// 	if instrT.ParamLen > 2 {
		// 		p.SetVarGlobal(pr, tk.ToBool(p.GetVarValue(r, instrT.Params[2])))
		// 	} else {
		// 		p.SetVarGlobal(pr, false)
		// 	}
		// } else if v1 == "int" {
		// 	if instrT.ParamLen > 2 {
		// 		p.SetVarGlobal(pr, tk.ToInt(p.GetVarValue(r, instrT.Params[2])))
		// 	} else {
		// 		p.SetVarGlobal(pr, int(0))
		// 	}
		// } else if v1 == "byte" {
		// 	if instrT.ParamLen > 2 {
		// 		p.SetVarGlobal(pr, tk.ToByte(p.GetVarValue(r, instrT.Params[2])))
		// 	} else {
		// 		p.SetVarGlobal(pr, byte(0))
		// 	}
		// } else if v1 == "rune" {
		// 	if instrT.ParamLen > 2 {
		// 		p.SetVarGlobal(pr, tk.ToRune(p.GetVarValue(r, instrT.Params[2])))
		// 	} else {
		// 		p.SetVarGlobal(pr, rune(0))
		// 	}
		// } else if v1 == "float" {
		// 	if instrT.ParamLen > 2 {
		// 		p.SetVarGlobal(pr, tk.ToFloat(p.GetVarValue(r, instrT.Params[2])))
		// 	} else {
		// 		p.SetVarGlobal(pr, float64(0.0))
		// 	}
		// } else if v1 == "str" {
		// 	if instrT.ParamLen > 2 {
		// 		p.SetVarGlobal(pr, tk.ToStr(p.GetVarValue(r, instrT.Params[2])))
		// 	} else {
		// 		p.SetVarGlobal(pr, "")
		// 	}
		// } else if v1 == "list" || v1 == "array" || v1 == "[]" {
		// 	blT := make([]interface{}, 0, instrT.ParamLen-2)

		// 	vs := p.ParamsToList(r, instrT, v1p+1)

		// 	for _, vvv := range vs {
		// 		nv, ok := vvv.([]interface{})

		// 		if ok {
		// 			for _, vvvj := range nv {
		// 				blT = append(blT, vvvj)
		// 			}
		// 		} else {
		// 			blT = append(blT, vvv)
		// 		}
		// 	}

		// 	p.SetVarGlobal(pr, blT)
		// } else if v1 == "strList" {
		// 	blT := make([]string, 0, instrT.ParamLen-2)

		// 	vs := p.ParamsToList(r, instrT, v1p+1)

		// 	for _, vvv := range vs {
		// 		nv, ok := vvv.([]string)

		// 		if ok {
		// 			for _, vvvj := range nv {
		// 				blT = append(blT, vvvj)
		// 			}
		// 		} else {
		// 			blT = append(blT, tk.ToStr(vvv))
		// 		}
		// 	}

		// 	p.SetVarGlobal(pr, blT)
		// } else if v1 == "byteList" {
		// 	blT := make([]byte, 0, instrT.ParamLen-2)

		// 	vs := p.ParamsToList(r, instrT, v1p+1)

		// 	for _, vvv := range vs {
		// 		nv, ok := vvv.([]byte)

		// 		if ok {
		// 			for _, vvvj := range nv {
		// 				blT = append(blT, vvvj)
		// 			}
		// 		} else {
		// 			blT = append(blT, tk.ToByte(vvv, 0))
		// 		}
		// 	}

		// 	p.SetVarGlobal(pr, blT)
		// } else if v1 == "runeList" {
		// 	blT := make([]rune, 0, instrT.ParamLen-2)

		// 	vs := p.ParamsToList(r, instrT, v1p+1)

		// 	for _, vvv := range vs {
		// 		nv, ok := vvv.([]rune)

		// 		if ok {
		// 			for _, vvvj := range nv {
		// 				blT = append(blT, vvvj)
		// 			}
		// 		} else {
		// 			blT = append(blT, tk.ToRune(vvv, 0))
		// 		}
		// 	}

		// 	p.SetVarGlobal(pr, blT)
		// } else if v1 == "map" {
		// 	p.SetVarGlobal(pr, map[string]interface{}{})
		// } else if v1 == "strMap" {
		// 	p.SetVarGlobal(pr, map[string]string{})
		// } else if v1 == "time" || v1 == "time.Time" {
		// 	if instrT.ParamLen > 2 {
		// 		p.SetVarGlobal(pr, tk.ToTime(p.GetVarValue(r, instrT.Params[2])))
		// 	} else {
		// 		p.SetVarGlobal(pr, time.Now())
		// 	}
		// } else {
		// 	switch v1 {
		// 	case "gui":
		// 		objT := p.GetVarValue(r, VarRef{Ref: 3, Value: "guiG"})
		// 		p.SetVarGlobal(pr, objT)
		// 	case "quickDelegate":
		// 		if instrT.ParamLen < 2 {
		// 			return p.Errf(r, "not enough parameters(参数不够)")
		// 		}

		// 		v2 := r.GetLabelIndex(p.GetVarValue(r, instrT.Params[v1p+1]))

		// 		var deleT tk.QuickDelegate

		// 		// same as fastCall, to be modified!!!
		// 		deleT = func(strA string) string {
		// 			pointerT := r.CodePointer

		// 			p.Stack.Push(strA)

		// 			tmpPointerT := v2

		// 			for {
		// 				rs := RunInstr(p, r, &r.InstrList[tmpPointerT])

		// 				nv, ok := rs.(int)

		// 				if ok {
		// 					tmpPointerT = nv
		// 					continue
		// 				}

		// 				nsv, ok := rs.(string)

		// 				if ok {
		// 					if tk.IsErrStr(nsv) {
		// 						// tmpRs := p.Pop()
		// 						r.CodePointer = pointerT
		// 						return nsv
		// 					}

		// 					if nsv == "exit" { // 不应发生
		// 						tmpRs := p.Stack.Pop()
		// 						r.CodePointer = pointerT
		// 						return tk.ToStr(tmpRs)
		// 					} else if nsv == "fr" {
		// 						break
		// 					}
		// 				}

		// 				tmpPointerT++
		// 			}

		// 			// return pointerT + 1

		// 			tmpRs := p.Stack.Pop()
		// 			r.CodePointer = pointerT
		// 			return tk.ToStr(tmpRs)
		// 		}

		// 		p.SetVarGlobal(pr, deleT)
		// 	case "image.Point", "point":
		// 		var p1 image.Point
		// 		if instrT.ParamLen > 3 {
		// 			p1 = image.Point{X: tk.ToInt(p.GetVarValue(r, instrT.Params[2])), Y: tk.ToInt(p.GetVarValue(r, instrT.Params[3]))}
		// 			p.SetVarGlobal(pr, p1)
		// 		} else {
		// 			p.SetVarGlobal(pr, p1)
		// 		}
		// 	default:
		// 		p.SetVarGlobal(pr, nil)

		// 	}

		// }

		// return ""

	case 203: // var
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough paramters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		// contextT := p.CurrentFuncContextM

		if instrT.ParamLen < 2 {
			p.SetVarLocal(r, pr, nil)
			// contextT.VarsM[nameT] = ""
			return ""
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		vs := p.ParamsToList(r, instrT, v1p+1)

		var pv interface{} = NewVar(p, r, v1, vs...)

		p.SetVarLocal(r, pr, pv)

		return ""
	case 205: // const
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		p.SetVar(r, pr, GlobalsG.Vars[tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))])

		return ""

	case 207: // nil
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// var pr interface{} = -5
		// v1p := 0

		// if instrT.ParamLen > 1 {
		pr := instrT.Params[0]
		// v1p := 1
		// }

		p.SetVar(r, pr, nil)

		return ""

	case 210: // ref
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		valueT := reflect.ValueOf(v1)

		if !valueT.CanAddr() {
			p.SetVar(r, pr, fmt.Errorf("not addressable(无法取引用)"))
			return ""
		}

		p.SetVar(r, pr, valueT.Addr().Interface())
		return ""

	case 215: // unref
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5

		var v2 interface{}

		if instrT.ParamLen < 1 {
			v2 = p.GetCurrentFuncContext(r).Tmp
		} else if instrT.ParamLen < 2 {
			v2 = p.GetVarValue(r, instrT.Params[0])
		} else {
			pr = instrT.Params[0]
			v2 = p.GetVarValue(r, instrT.Params[1])
		}

		switch nv := v2.(type) {
		case *interface{}:
			p.SetVar(r, pr, *nv)
		case *byte:
			p.SetVar(r, pr, *nv)
		case *int:
			p.SetVar(r, pr, *nv)
		case *uint64:
			p.SetVar(r, pr, *nv)
		case *rune:
			p.SetVar(r, pr, *nv)
		case *bool:
			p.SetVar(r, pr, *nv)
		case *string:
			p.SetVar(r, pr, *nv)
		case *strings.Builder:
			p.SetVar(r, pr, *nv)
		default:
			valueT := reflect.ValueOf(v2)
			kindT := valueT.Kind()

			if kindT == reflect.Pointer {
				p.SetVar(r, pr, valueT.Elem().Interface())
				return ""
			}

			return p.Errf(r, "unsupport type(无法处理的类型): %T", v2)
		}

		return ""

	// case 210: // ref
	// 	if instrT.ParamLen < 1 {
	// 		return p.Errf(r, "not enough parameters(参数不够)")
	// 	}

	// 	var pr interface{} = -5

	// 	var v2 interface{}

	// 	if instrT.ParamLen < 2 {
	// 		v2 = p.GetVarRef(instrT.Params[0])
	// 	} else {
	// 		pr = instrT.Params[0]
	// 		v2 = p.GetVarRef(instrT.Params[1])
	// 	}

	// 	p.SetVar(r, pr, v2)

	// 	return ""
	// case 211: // refNative
	// 	if instrT.ParamLen < 1 {
	// 		return p.Errf(r, "not enough parameters(参数不够)")
	// 	}

	// 	var pr interface{} = -5

	// 	var v2 interface{}

	// 	if instrT.ParamLen < 2 {
	// 		v2 = p.GetVarRefNative(instrT.Params[0])
	// 	} else {
	// 		pr = instrT.Params[0]
	// 		v2 = p.GetVarRefNative(instrT.Params[1])
	// 	}

	// 	p.SetVar(r, pr, v2)

	// 	return ""
	// case 215: // unref
	// 	if instrT.ParamLen < 1 {
	// 		return p.Errf(r, "not enough parameters(参数不够)")
	// 	}

	// 	var pr interface{} = -5

	// 	var v2 interface{}

	// 	if instrT.ParamLen < 1 {
	// 		v2 = p.GetCurrentFuncContext(r).Tmp
	// 	} else if instrT.ParamLen < 2 {
	// 		v2 = p.GetVarValue(r, instrT.Params[0])
	// 	} else {
	// 		pr = instrT.Params[0]
	// 		v2 = p.GetVarValue(r, instrT.Params[1])
	// 	}

	// 	switch nv := v2.(type) {
	// 	case *interface{}:
	// 		p.SetVar(r, pr, *nv)
	// 	case *byte:
	// 		p.SetVar(r, pr, *nv)
	// 	case *int:
	// 		p.SetVar(r, pr, *nv)
	// 	case *uint64:
	// 		p.SetVar(r, pr, *nv)
	// 	case *rune:
	// 		p.SetVar(r, pr, *nv)
	// 	case *bool:
	// 		p.SetVar(r, pr, *nv)
	// 	case *string:
	// 		p.SetVar(r, pr, *nv)
	// 	case *strings.Builder:
	// 		p.SetVar(r, pr, *nv)
	// 	default:
	// 		return p.Errf(r, "无法处理的类型：%T", v2)
	// 	}

	// 	return ""
	case 218: // assignRef
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// nameT := instrT.Params[0].Ref
		p1a := p.GetVarValue(r, instrT.Params[0])

		// p1, ok := p1a.(*interface{})

		// var p1 *interface{}
		v1p := 1
		v1 := p.GetVarValue(r, instrT.Params[v1p])

		switch nv := p1a.(type) {
		case *interface{}:
			*nv = v1
			return ""
			// p1 = nv
			// break
		case *byte:
			*nv = tk.ToByte(v1)
			return ""
		case *rune:
			*nv = tk.ToRune(v1)
			return ""
		case *int:
			*nv = tk.ToInt(v1)
			return ""
		default:
			errT := tk.SetByRef(p1a, v1)

			if errT != nil {
				return p.Errf(r, "failed to set value by ref(按引用赋值失败): %v -> %T(%v)", errT, p1a, p1a)
			}

			return ""
		}

		// if instrT.ParamLen > 2 {
		// 	valueTypeT := instrT.Params[v1p].Value
		// 	valueT := p.GetVarValue(r, instrT.Params[v1p+1])

		// 	if valueTypeT == "bool" {
		// 		*p1 = tk.ToBool(valueT)
		// 	} else if valueTypeT == "int" {
		// 		*p1 = tk.ToInt(valueT)
		// 	} else if valueTypeT == "byte" {
		// 		*p1 = tk.ToByte(valueT)
		// 	} else if valueTypeT == "rune" {
		// 		*p1 = tk.ToRune(valueT)
		// 	} else if valueTypeT == "float" {
		// 		*p1 = tk.ToFloat(valueT)
		// 	} else if valueTypeT == "str" {
		// 		*p1 = tk.ToStr(valueT)
		// 	} else if valueTypeT == "list" || valueT == "array" || valueT == "[]" {
		// 		*p1 = valueT.([]interface{})
		// 	} else if valueTypeT == "strList" {
		// 		*p1 = valueT.([]string)
		// 	} else if valueTypeT == "byteList" {
		// 		*p1 = valueT.([]byte)
		// 	} else if valueTypeT == "runeList" {
		// 		*p1 = valueT.([]rune)
		// 	} else if valueTypeT == "map" {
		// 		*p1 = valueT.(map[string]interface{})
		// 	} else if valueTypeT == "strMap" {
		// 		*p1 = valueT.(map[string]string)
		// 	} else if valueTypeT == "time" {
		// 		*p1 = valueT.(time.Time)
		// 	} else {
		// 		*p1 = valueT
		// 	}

		// 	return ""
		// }

		// valueT := p.GetVarValue(r, instrT.Params[1])

		// *p1 = valueT

		// // (*(p.CurrentVarsM))[nameT] = valueT

		// return ""

	case 220: // push
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		if instrT.ParamLen > 1 {
			v2 := p.GetVarValue(r, instrT.Params[1])

			v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[0]))

			if v1 == "int" {
				p.Stack.Push(tk.ToInt(v2))
			} else if v1 == "byte" {
				p.Stack.Push(tk.ToByte(v2))
			} else if v1 == "rune" {
				p.Stack.Push(tk.ToRune(v2))
			} else if v1 == "float" {
				p.Stack.Push(tk.ToFloat(v2))
			} else if v1 == "bool" {
				p.Stack.Push(tk.ToBool(v2))
			} else if v1 == "str" {
				p.Stack.Push(tk.ToStr(v2))
			} else {
				p.Stack.Push(v2)
			}

			return ""
		}

		v1 := p.GetVarValue(r, instrT.Params[0])

		p.Stack.Push(v1)

		return ""
	case 222: // peek
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		errT := p.SetVar(r, pr, p.Stack.Peek())

		if errT != nil {
			return p.Errf(r, "%v", errT)
		}

		return ""

	case 224: // pop
		var pr interface{} = -5
		if instrT.ParamLen < 1 {
			p.SetVar(r, pr, p.Stack.Pop())
			return ""
		}

		pr = instrT.Params[0]

		p.SetVar(r, pr, p.Stack.Pop())

		return ""
	case 230: // getStackSize
		var pr interface{} = -5
		if instrT.ParamLen < 1 {
			p.SetVar(r, pr, p.Stack.Size())
			return ""
		}

		pr = instrT.Params[0]

		p.SetVar(r, pr, p.Stack.Size())

		return ""
	case 240: // clearStack
		p.Stack.Clear()
		return ""

	case 250: // pushRun
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		if instrT.ParamLen > 1 {
			v2 := p.GetVarValue(r, instrT.Params[1])

			v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[0]))

			if v1 == "int" {
				r.PointerStack.Push(tk.ToInt(v2))
			} else if v1 == "byte" {
				r.PointerStack.Push(tk.ToByte(v2))
			} else if v1 == "rune" {
				r.PointerStack.Push(tk.ToRune(v2))
			} else if v1 == "float" {
				r.PointerStack.Push(tk.ToFloat(v2))
			} else if v1 == "bool" {
				r.PointerStack.Push(tk.ToBool(v2))
			} else if v1 == "str" {
				r.PointerStack.Push(tk.ToStr(v2))
			} else {
				r.PointerStack.Push(v2)
			}

			return ""
		}

		v1 := p.GetVarValue(r, instrT.Params[0])

		r.PointerStack.Push(v1)

		return ""
	case 252: // peekRun
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		errT := p.SetVar(r, pr, r.PointerStack.Peek())

		if errT != nil {
			return p.Errf(r, "%v", errT)
		}

		return ""

	case 254: // popRun
		var pr interface{} = -5
		if instrT.ParamLen < 1 {
			p.SetVar(r, pr, r.PointerStack.Pop())
			return ""
		}

		pr = instrT.Params[0]

		p.SetVar(r, pr, r.PointerStack.Pop())

		return ""
	case 256: // getRunStackSize
		var pr interface{} = -5
		if instrT.ParamLen < 1 {
			p.SetVar(r, pr, r.PointerStack.Size())
			return ""
		}

		pr = instrT.Params[0]

		p.SetVar(r, pr, r.PointerStack.Size())

		return ""
	case 258: // clearRunStack
		r.PointerStack.Clear()
		return ""

	case 300: // getSharedMap

		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		p.SetVar(r, pr, GlobalsG.SyncMap.GetList())

		return ""

	case 301: // getSharedMapItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0] // -5
		v1p := 1

		// if instrT.ParamLen > 2 {
		// 	pr = instrT.Params[0]
		// 	v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		// v1 := p.GetVarValue(r, instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		// var vT interface{}
		// tk.Pln(pr, v1, v2)

		var defaultT interface{} = tk.Undefined

		if instrT.ParamLen > 2 {
			defaultT = p.GetVarValue(r, instrT.Params[2])
		}

		p.SetVar(r, pr, GlobalsG.SyncMap.Get(v1, defaultT))

		return ""

	case 302: // getSharedMapSize

		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		p.SetVar(r, pr, GlobalsG.SyncMap.Size())

		return ""

	case 303: // tryGetSharedMapItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0] // -5
		v1p := 1

		// if instrT.ParamLen > 2 {
		// 	pr = instrT.Params[0]
		// 	v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		// v1 := p.GetVarValue(r, instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		// var vT interface{}
		// tk.Pln(pr, v1, v2)

		var defaultT interface{} = tk.Undefined

		if instrT.ParamLen > 2 {
			defaultT = p.GetVarValue(r, instrT.Params[2])
		}

		p.SetVar(r, pr, GlobalsG.SyncMap.TryGet(v1, defaultT))

		return ""

	case 304: // tryGetSharedMapSize
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		vr := GlobalsG.SyncMap.TrySize()

		if vr < 0 {
			p.SetVar(r, pr, fmt.Errorf("获取大小失败"))
		} else {
			p.SetVar(r, pr, vr)
		}

		return ""

	case 311: // setSharedMapItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// var pr interface{} = -5
		v1p := 0

		// if instrT.ParamLen > 2 {
		// 	// pr = instrT.Params[0]
		// 	v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		// v1 := p.GetVarValue(r, instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		GlobalsG.SyncMap.Set(v1, v2)
		// p.SetVar(r, pr, true)

		return ""

	case 313: // trySetSharedMapItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
			// p.SetVarInt(instrT.Params[2].Ref, vT)
		}

		// v1 := p.GetVarValue(r, instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		p.SetVar(r, pr, GlobalsG.SyncMap.TrySet(v1, v2))

		return ""

	case 321: // deleteSharedMapItem
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// var pr interface{} = -5
		v1p := 0

		// if instrT.ParamLen > 1 {
		// 	// pr = instrT.Params[0]
		// 	v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		// v1 := p.GetVarValue(r, instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		GlobalsG.SyncMap.Delete(v1)
		// p.SetVar(r, pr, true)

		return ""

	case 323: // tryDeleteSharedMapItem
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
			// p.SetVarInt(instrT.Params[2].Ref, vT)
		}

		// v1 := p.GetVarValue(r, instrT.Params[v1p])

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, GlobalsG.SyncMap.TryDelete(v1))

		return ""

	case 331: // clearSharedMapItem
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		// var pr interface{} = -5
		// // v1p := 0

		// if instrT.ParamLen > 1 {
		// 	pr = instrT.Params[0]
		// 	// v1p = 1
		// 	// p.SetVarInt(instrT.Params[2].Ref, vT)
		// }

		// v1 := p.GetVarValue(r, instrT.Params[v1p])

		// v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		GlobalsG.SyncMap.Clear()
		// p.SetVar(r, pr, true)

		return ""

	case 333: // tryClearSharedMap
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			// v1p = 1
			// p.SetVarInt(instrT.Params[2].Ref, vT)
		}

		// v1 := p.GetVarValue(r, instrT.Params[v1p])

		// v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, GlobalsG.SyncMap.TryClear())

		return ""

	case 341: // lockSharedMap
		GlobalsG.SyncMap.Lock()

		return ""

	case 342: // tryLockSharedMap
		GlobalsG.SyncMap.TryLock()

		return ""

	case 343: // unlockSharedMap
		GlobalsG.SyncMap.Unlock()

		return ""

	case 346: // readLockSharedMap
		GlobalsG.SyncMap.RLock()

		return ""

	case 347: // tryReadLockSharedMap
		GlobalsG.SyncMap.TryRLock()

		return ""

	case 348: // readUnlockSharedMap
		GlobalsG.SyncMap.RUnlock()

		return ""

	case 351: // quickClearSharedMap
		GlobalsG.SyncMap.QuickClear()

		return ""

	case 353: // quickGetSharedMapItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0] // -5
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		var defaultT interface{} = tk.Undefined

		if instrT.ParamLen > 2 {
			defaultT = p.GetVarValue(r, instrT.Params[2])
		}

		p.SetVar(r, pr, GlobalsG.SyncMap.QuickGet(v1, defaultT))

		return ""

	case 354: // quickGetSharedMap
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		p.SetVar(r, pr, GlobalsG.SyncMap)

		return ""

	case 355: // quickSetSharedMapItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1p := 0

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		GlobalsG.SyncMap.QuickSet(v1, v2)

		return ""

	case 357: // quickDeleteSharedMapItem
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1p := 0

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		GlobalsG.SyncMap.QuickDelete(v1)

		return ""
	case 359: // quickSizeSharedMap

		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		p.SetVar(r, pr, GlobalsG.SyncMap.QuickSize())

		return ""

	case 401: // assign/=
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		var valueT interface{}

		if instrT.ParamLen > 2 {
			valueTypeT := instrT.Params[1].Value
			valueT = p.GetVarValue(r, instrT.Params[2])

			if valueTypeT == "bool" {
				valueT = tk.ToBool(valueT)
			} else if valueTypeT == "int" {
				valueT = tk.ToInt(valueT)
				// p.SetVar(r, pr, tk.ToInt(valueT))
			} else if valueTypeT == "byte" {
				valueT = tk.ToByte(valueT)
			} else if valueTypeT == "rune" {
				valueT = tk.ToRune(valueT)
			} else if valueTypeT == "float" {
				valueT = tk.ToFloat(valueT)
			} else if valueTypeT == "str" {
				valueT = tk.ToStr(valueT)
				// } else if valueTypeT == "list" || valueT == "array" || valueT == "[]" {
				// 	p.SetVar(r, pr, valueT.([]interface{}))
				// } else if valueTypeT == "strList" {
				// 	p.SetVar(r, pr, valueT.([]string))
				// } else if valueTypeT == "byteList" {
				// 	p.SetVar(r, pr, valueT.([]byte))
				// } else if valueTypeT == "runeList" {
				// 	p.SetVar(r, pr, valueT.([]rune))
				// } else if valueTypeT == "map" {
				// 	p.SetVar(r, pr, valueT.(map[string]interface{}))
				// } else if valueTypeT == "strMap" {
				// 	p.SetVar(r, pr, valueT.(map[string]string))
				// } else if valueTypeT == "time" {
				// 	p.SetVar(r, pr, valueT.(map[string]string))
			}
		} else {
			valueT = p.GetVarValue(r, instrT.Params[1])
		}

		p.SetVar(r, pr, valueT)

		return ""

	case 491: // assignGlobal
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		var valueT interface{}

		if instrT.ParamLen > 2 {
			valueTypeT := instrT.Params[1].Value
			valueT = p.GetVarValue(r, instrT.Params[2])

			if valueTypeT == "bool" {
				valueT = tk.ToBool(valueT)
			} else if valueTypeT == "int" {
				valueT = tk.ToInt(valueT)
				// p.SetVar(r, pr, tk.ToInt(valueT))
			} else if valueTypeT == "byte" {
				valueT = tk.ToByte(valueT)
			} else if valueTypeT == "rune" {
				valueT = tk.ToRune(valueT)
			} else if valueTypeT == "float" {
				valueT = tk.ToFloat(valueT)
			} else if valueTypeT == "str" {
				valueT = tk.ToStr(valueT)
			}
		} else {
			valueT = p.GetVarValue(r, instrT.Params[1])
		}

		p.SetVarGlobal(pr, valueT)

		return ""

	case 492: // assignFromGlobal
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		var valueT interface{}

		if instrT.ParamLen > 2 {
			valueTypeT := instrT.Params[1].Value
			valueT = p.GetVarValueGlobal(r, instrT.Params[2])

			if valueTypeT == "bool" {
				valueT = tk.ToBool(valueT)
			} else if valueTypeT == "int" {
				valueT = tk.ToInt(valueT)
				// p.SetVar(r, pr, tk.ToInt(valueT))
			} else if valueTypeT == "byte" {
				valueT = tk.ToByte(valueT)
			} else if valueTypeT == "rune" {
				valueT = tk.ToRune(valueT)
			} else if valueTypeT == "float" {
				valueT = tk.ToFloat(valueT)
			} else if valueTypeT == "str" {
				valueT = tk.ToStr(valueT)
			}
		} else {
			valueT = p.GetVarValueGlobal(r, instrT.Params[1])
		}

		p.SetVarGlobal(pr, valueT)

		return ""

	case 493: // assignLocal
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		var valueT interface{}

		if instrT.ParamLen > 2 {
			valueTypeT := instrT.Params[1].Value
			valueT = p.GetVarValue(r, instrT.Params[2])

			if valueTypeT == "bool" {
				valueT = tk.ToBool(valueT)
			} else if valueTypeT == "int" {
				valueT = tk.ToInt(valueT)
				// p.SetVar(r, pr, tk.ToInt(valueT))
			} else if valueTypeT == "byte" {
				valueT = tk.ToByte(valueT)
			} else if valueTypeT == "rune" {
				valueT = tk.ToRune(valueT)
			} else if valueTypeT == "float" {
				valueT = tk.ToFloat(valueT)
			} else if valueTypeT == "str" {
				valueT = tk.ToStr(valueT)
			}
		} else {
			valueT = p.GetVarValue(r, instrT.Params[1])
		}

		p.SetVarLocal(r, pr, valueT)

		return ""

	case 610: // if
		// tk.Plv(instrT)
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var condT bool
		var v2 interface{}
		var v2o interface{}
		var ok0 bool

		var elseLabelIntT int = -1

		if instrT.ParamLen > 2 {
			elseLabelT := r.GetLabelIndex(p.GetVarValue(r, instrT.Params[2]))

			if elseLabelT < 0 {
				return p.Errf(r, "invalid label: %v", elseLabelT)
			}

			elseLabelIntT = elseLabelT
		}

		v2o = instrT.Params[1]

		v2 = p.GetVarValue(r, instrT.Params[1])

		// tk.Plv(instrT)
		tmpv := p.GetVarValue(r, instrT.Params[0])
		if GlobalsG.VerboseLevel > 1 {
			tk.Pl("if %v -> %v", instrT.Params[0], tmpv)
		}

		condT, ok0 = tmpv.(bool)

		if !ok0 {
			var tmps string
			tmps, ok0 = tmpv.(string)

			if ok0 {
				tmprs := QuickEval(tmps, p, r)

				condT, ok0 = tmprs.(bool)
			}
		}

		if !ok0 {
			return p.Errf(r, "invalid condition parameter: %#v", tmpv)
		}

		if condT {
			c2 := r.GetLabelIndex(v2)

			if c2 < 0 {
				return p.Errf(r, "invalid label: %v", v2o)
			}

			return c2
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""
	case 611: // ifNot
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var condT bool
		var v2 interface{}
		var v2o interface{}
		var ok0 bool

		var elseLabelIntT int = -1

		if instrT.ParamLen > 2 {
			elseLabelT := r.GetLabelIndex(p.GetVarValue(r, instrT.Params[2]))

			if elseLabelT < 0 {
				return p.Errf(r, "invalid label: %v", elseLabelT)
			}

			elseLabelIntT = elseLabelT
		}

		v2o = instrT.Params[1]

		v2 = p.GetVarValue(r, instrT.Params[1])

		// tk.Plv(instrT)
		tmpv := p.GetVarValue(r, instrT.Params[0])
		if GlobalsG.VerboseLevel > 1 {
			tk.Pl("if %v -> %v", instrT.Params[0], tmpv)
		}

		condT, ok0 = tmpv.(bool)

		if !ok0 {
			var tmps string
			tmps, ok0 = tmpv.(string)

			if ok0 {
				tmprs := QuickEval(tmps, p, r)

				condT, ok0 = tmprs.(bool)
			}
		}

		if !ok0 {
			return p.Errf(r, "invalid condition parameter: %#v", tmpv)
		}

		if !condT {
			c2 := r.GetLabelIndex(v2)

			if c2 < 0 {
				return p.Errf(r, "invalid label: %v", v2o)
			}

			return c2
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 631: // ifEval
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var condT bool
		var v2 interface{}
		var v2o interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 2 {
			elseLabelT := r.GetLabelIndex(p.GetVarValue(r, instrT.Params[2]))

			if elseLabelT < 0 {
				return p.Errf(r, "invalid label: %v", elseLabelT)
			}

			elseLabelIntT = elseLabelT
		}

		v2o = instrT.Params[1]

		v2 = p.GetVarValue(r, instrT.Params[1])

		// tk.Plv(instrT)
		tmpv := tk.ToStr(p.GetVarValue(r, instrT.Params[0]))

		condObjT := QuickEval(tmpv, p, r)

		if tk.IsError(condObjT) {
			return p.Errf(r, "failed to compare: %v", condObjT)
		}

		condT = tk.ToBool(condObjT)

		if condT {
			c2 := r.GetLabelIndex(v2)

			if c2 < 0 {
				return p.Errf(r, "invalid label: %v", v2o)
			}

			return c2
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 641: // ifEmpty
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var condT bool
		var v2 interface{}
		var v2o interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 2 {
			elseLabelT := r.GetLabelIndex(p.GetVarValue(r, instrT.Params[2]))

			if elseLabelT < 0 {
				return p.Errf(r, "invalid label: %v", elseLabelT)
			}

			elseLabelIntT = elseLabelT
		}

		v2o = instrT.Params[1]

		v2 = p.GetVarValue(r, instrT.Params[1])

		// tk.Plv(instrT)
		v1 := p.GetVarValue(r, instrT.Params[0])

		if v1 == nil {
			condT = true
		} else if v1 == tk.Undefined {
			condT = true
		} else {
			switch nv := v1.(type) {
			// case bool:
			// 	condT = (nv == false)
			case string:
				condT = (nv == "")
			// case byte:
			// 	condT = (nv <= 0)
			// case int:
			// 	condT = (nv <= 0)
			// case rune:
			// 	condT = (nv <= 0)
			// case int64:
			// 	condT = (nv <= 0)
			// case float64:
			// 	condT = (nv <= 0)
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

		if condT {
			c2 := r.GetLabelIndex(v2)

			if c2 < 0 {
				return p.Errf(r, "invalid label: %v", v2o)
			}

			return c2
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 643: // ifEqual
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var condT bool
		var v2 interface{}
		var v2o interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 3 {
			elseLabelT := r.GetLabelIndex(p.GetVarValue(r, instrT.Params[3]))

			if elseLabelT < 0 {
				return p.Errf(r, "invalid label: %v", elseLabelT)
			}

			elseLabelIntT = elseLabelT
		}

		v2o = instrT.Params[2]

		v2 = p.GetVarValue(r, instrT.Params[2])

		// tk.Plv(instrT)
		tmpv := p.GetVarValue(r, instrT.Params[0])
		tmpv2 := p.GetVarValue(r, instrT.Params[1])

		condObjT := tk.GetEQResult(tmpv, tmpv2)

		if tk.IsError(condObjT) {
			return p.Errf(r, "failed to compare: %v", condObjT)
		}

		condT = tk.ToBool(condObjT)

		if condT {
			c2 := r.GetLabelIndex(v2)

			if c2 < 0 {
				return p.Errf(r, "invalid label: %v", v2o)
			}

			return c2
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 644: // ifNotEqual
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var condT bool
		var v2 interface{}
		var v2o interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 3 {
			elseLabelT := r.GetLabelIndex(p.GetVarValue(r, instrT.Params[3]))

			if elseLabelT < 0 {
				return p.Errf(r, "invalid label: %v", elseLabelT)
			}

			elseLabelIntT = elseLabelT
		}

		v2o = instrT.Params[2]

		v2 = p.GetVarValue(r, instrT.Params[2])

		// tk.Plv(instrT)
		tmpv := p.GetVarValue(r, instrT.Params[0])
		tmpv2 := p.GetVarValue(r, instrT.Params[1])

		condObjT := tk.GetNEQResult(tmpv, tmpv2)

		if tk.IsError(condObjT) {
			return p.Errf(r, "failed to compare: %v", condObjT)
		}

		condT = tk.ToBool(condObjT)

		if condT {
			c2 := r.GetLabelIndex(v2)

			if c2 < 0 {
				return p.Errf(r, "invalid label: %v", v2o)
			}

			return c2
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 651: // ifErr/IfErrX
		// tk.Plv(instrT)
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var condT bool
		var v2 interface{}
		var v2o interface{}

		var elseLabelIntT int = -1

		if instrT.ParamLen > 2 {
			elseLabelT := r.GetLabelIndex(p.GetVarValue(r, instrT.Params[2]))

			if elseLabelT < 0 {
				return p.Errf(r, "invalid label: %v", elseLabelT)
			}

			elseLabelIntT = elseLabelT
		}

		v2o = instrT.Params[1]

		v2 = p.GetVarValue(r, instrT.Params[1])

		// tk.Plv(instrT)
		tmpv := p.GetVarValue(r, instrT.Params[0])
		if GlobalsG.VerboseLevel > 1 {
			tk.Pl("if %v -> %v", instrT.Params[0], tmpv)
		}

		condT = tk.IsErrX(tmpv)

		if condT {
			c2 := r.GetLabelIndex(v2)

			if c2 < 0 {
				return p.Errf(r, "invalid label: %v", v2o)
			}

			return c2
		}

		if elseLabelIntT >= 0 {
			return elseLabelIntT
		}

		return ""

	case 691: // switch
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0])

		vs := p.ParamsToList(r, instrT, 1)

		len1T := len(vs)

		lenT := len1T / 2

		for i := 0; i < lenT; i++ {
			if v1 == vs[i*2] {
				labelT := vs[i*2+1]

				// tk.Pl("labelT: %v", labelT)

				c := r.GetLabelIndex(labelT)

				// c, ok := labelT.(int)

				if c >= 0 {
					return c
				} else {
					return p.Errf(r, "标号格式错误：%T(%v)", labelT, labelT)
				}
			}
		}

		if len1T > (lenT * 2) {
			labelT := vs[len1T-1]

			c := r.GetLabelIndex(labelT)

			if c >= 0 {
				return c
			} else {
				return p.Errf(r, "标号格式错误：%T(%v)", labelT, labelT)
			}

		}

		return ""

	case 693: // switchCond
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		vs := p.ParamsToList(r, instrT, 0)

		len1T := len(vs)

		lenT := len1T / 2

		for i := 0; i < lenT; i++ {
			objT := vs[i*2]

			nv, ok := objT.(string)
			if ok {
				objT = QuickEval(nv, p, r)
			}

			if tk.ToBool(objT) {
				labelT := vs[i*2+1]

				// tk.Pl("labelT: %v", labelT)

				c := r.GetLabelIndex(labelT)

				// c, ok := labelT.(int)

				if c >= 0 {
					return c
				} else {
					return p.Errf(r, "标号格式错误：%T(%v)", labelT, labelT)
				}
			}
		}

		if len1T > (lenT * 2) {
			labelT := vs[len1T-1]

			c := r.GetLabelIndex(labelT)

			if c >= 0 {
				return c
			} else {
				return p.Errf(r, "标号格式错误：%T(%v)", labelT, labelT)
			}

		}

		return ""

	case 701: // ==
		var pr interface{} = -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0]
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(r, instrT.Params[0])
			v2 = p.GetVarValue(r, instrT.Params[1])
		} else {
			pr = instrT.Params[0]
			v1 = p.GetVarValue(r, instrT.Params[1])
			v2 = p.GetVarValue(r, instrT.Params[2])
		}

		v3 := tk.GetEQResult(v1, v2)

		p.SetVar(r, pr, v3)
		return ""

	case 702: // !=
		var pr interface{} = -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0]
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(r, instrT.Params[0])
			v2 = p.GetVarValue(r, instrT.Params[1])
		} else {
			pr = instrT.Params[0]
			v1 = p.GetVarValue(r, instrT.Params[1])
			v2 = p.GetVarValue(r, instrT.Params[2])
		}

		v3 := tk.GetNEQResult(v1, v2)

		p.SetVar(r, pr, v3)
		return ""

	case 703: // <
		var pr interface{} = -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0]
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(r, instrT.Params[0])
			v2 = p.GetVarValue(r, instrT.Params[1])
		} else {
			pr = instrT.Params[0]
			v1 = p.GetVarValue(r, instrT.Params[1])
			v2 = p.GetVarValue(r, instrT.Params[2])
		}

		v3 := tk.GetLTResult(v1, v2)

		p.SetVar(r, pr, v3)
		return ""

	case 704: // >
		var pr interface{} = -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0]
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(r, instrT.Params[0])
			v2 = p.GetVarValue(r, instrT.Params[1])
		} else {
			pr = instrT.Params[0]
			v1 = p.GetVarValue(r, instrT.Params[1])
			v2 = p.GetVarValue(r, instrT.Params[2])
		}

		v3 := tk.GetGTResult(v1, v2)

		p.SetVar(r, pr, v3)
		return ""

	case 705: // <=
		var pr interface{} = -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0]
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(r, instrT.Params[0])
			v2 = p.GetVarValue(r, instrT.Params[1])
		} else {
			pr = instrT.Params[0]
			v1 = p.GetVarValue(r, instrT.Params[1])
			v2 = p.GetVarValue(r, instrT.Params[2])
		}

		v3 := tk.GetLETResult(v1, v2)

		p.SetVar(r, pr, v3)
		return ""

	case 706: // >=
		var pr interface{} = -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0]
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(r, instrT.Params[0])
			v2 = p.GetVarValue(r, instrT.Params[1])
		} else {
			pr = instrT.Params[0]
			v1 = p.GetVarValue(r, instrT.Params[1])
			v2 = p.GetVarValue(r, instrT.Params[2])
		}

		v3 := tk.GetGETResult(v1, v2)

		p.SetVar(r, pr, v3)
		return ""
	case 790: // cmp/比较
		var pr interface{} = -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0]
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
			// return p.Errf(r, "not enough parameters(参数不够)")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(r, instrT.Params[0])
			v2 = p.GetVarValue(r, instrT.Params[1])
		} else {
			pr = instrT.Params[0]
			v1 = p.GetVarValue(r, instrT.Params[1])
			v2 = p.GetVarValue(r, instrT.Params[2])
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
			return p.Errf(r, "types not match(数据类型不匹配): %T(%#v） -> %T(%#v)", v1, v1, v2, v2)
		}

		p.SetVar(r, pr, v3)

		return ""

	case 801: // inc
		var pr interface{} = -5
		v1p := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			v1p = 0
		}

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		nv, ok := v1.(int)

		if ok {
			p.SetVar(r, pr, nv+1)
			return ""
		}

		nv2, ok := v1.(byte)

		if ok {
			p.SetVar(r, pr, nv2+1)
			return ""
		}

		nv3, ok := v1.(rune)

		if ok {
			p.SetVar(r, pr, nv3+1)
			return ""
		}

		p.SetVar(r, pr, tk.ToInt(v1)+1)

		return ""

	case 810: // dec
		var pr interface{} = -5
		v1p := -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			v1p = 0
		}

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		nv, ok := v1.(int)

		if ok {
			p.SetVar(r, pr, nv-1)
			return ""
		}

		nv2, ok := v1.(byte)

		if ok {
			p.SetVar(r, pr, nv2-1)
			return ""
		}

		nv3, ok := v1.(rune)

		if ok {
			p.SetVar(r, pr, nv3-1)
			return ""
		}

		p.SetVar(r, pr, tk.ToInt(v1)-1)

		return ""
	case 901: // add/+
		// tk.Plv(instrT)
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetAddResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""

	case 902: // sub/-
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetMinusResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""

	case 903: // mul/*
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetMultiplyResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""

	case 904: // div/"/"
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetDivResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""

	case 905: // mod/%
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetModResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""
	case 921: // adds
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		vs := p.ParamsToList(r, instrT, v1p)

		p.SetVar(r, pr, tk.GetAddsResult(vs...))

		return ""

	case 930: // !
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v3 := tk.GetLogicalNotResult(v1)

		p.SetVar(r, pr, v3)

		return ""

	case 931: // not
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v3 := tk.GetBitNotResult(v1)

		p.SetVar(r, pr, v3)

		return ""

	case 933: // &&
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetANDResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""

	case 934: // ||
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetORResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""

	case 941: // &
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetBitANDResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""

	case 942: // |
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetBitORResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""

	case 943: // ^/xor
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetBitXORResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""

	case 944: // &^/andNot
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := tk.GetBitANDNOTResult(v1, v2)

		p.SetVar(r, pr, v3)

		return ""
	// case 998: // eval
	// if instrT.ParamLen < 1 {
	// 	return p.Errf(r, "not enough parameters(参数个数不够)")
	// }

	// var pr interface{} = -5
	// v1p := 0

	// if instrT.ParamLen > 1 {
	// 	pr = instrT.Params[0]
	// 	v1p = 1
	// }

	// v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

	// p.SetVar(r, pr, p.EvalExpression(v1))

	// return ""

	case 990: // ? 三元操作符
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数个数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToBool(p.GetVarValue(r, instrT.Params[v1p]))

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := p.GetVarValue(r, instrT.Params[v1p+2])

		if v1 {
			p.SetVar(r, pr, v2)
		} else {
			p.SetVar(r, pr, v3)
		}

		return ""

	case 999: // eval/quickEval
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough paramters")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		// tk.Plv(instrT)

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])) //instrT.Line

		p.SetVar(r, pr, QuickEval(v1, p, r))

		return ""

	case 1010: // call
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough paramters")
		}

		pr := instrT.Params[0]

		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v1c := r.GetLabelIndex(v1)

		if v1c < 0 {
			return p.Errf(r, "invalid label format: %v", v1)
		}

		r.PointerStack.Push(CallStruct{Type: 2, ReturnPointer: r.CodePointer, ReturnRef: pr})

		r.PushFunc()

		if instrT.ParamLen > 2 {
			vs := p.ParamsToList(r, instrT, 2)

			r.GetCurrentFuncContext().Vars["inputL"] = vs
		}

		return v1c

	case 1020: // ret
		rs := r.PointerStack.Pop()

		if tk.IsUndefined(rs) {
			return p.Errf(r, "pointer stack empty")
		}

		nv, ok := rs.(CallStruct)

		if !ok {
			return p.Errf(r, "not in a call, not a call struct in running stack: %v", rs)
		}

		if nv.Type == 1 { // fastCall
			return tk.ToInt(nv.ReturnPointer) + 1
		} else if nv.Type == 2 { // (normal) call
			rsi := RunDefer(p, r)

			if tk.IsError(rsi) {
				return p.Errf(r, "[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), rsi)
			}

			if instrT.ParamLen > 0 {
				// tk.Pl("outL <-: %#v", p.GetVarValue(r, instrT.Params[0]))
				r.GetCurrentFuncContext().Vars["outL"] = p.GetVarValue(r, instrT.Params[0])
			}

			rs2, rok := r.GetCurrentFuncContext().Vars["outL"]

			errT := r.PopFunc()

			if errT != nil {
				return p.Errf(r, "failed to return from function call while popFunc: %v", errT)
			}

			pr := nv.ReturnRef

			if rok {
				p.SetVar(r, pr, rs2)
			} else {
				p.SetVar(r, pr, tk.Undefined)
			}

			// newPointT := r.PointerStack.Pop()

			// if newPointT == nil || tk.IsUndefined(newPointT) {
			// 	return p.Errf(r, "no return pointer from function call: %v", newPointT)
			// }

			return nv.ReturnPointer + 1
		}

		return p.Errf(r, "unsupported call type to return: %v", nv.Type)
	case 1050: // sealCall
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		codeT := p.GetVarValue(r, instrT.Params[v1p])

		if r1, ok := codeT.(*RunningContext); ok {
			vs := p.ParamsToList(r, instrT, v1p+1)
			rs := RunCodePiece(NewVMQuick(), r1, "", vs, true)

			p.SetVar(r, pr, rs)
			return ""
		}

		if c1, ok := codeT.(int); ok {
			if instrT.ParamLen < 3 {
				return p.Errf(r, "not enough parameters(参数不够)")
			}

			c2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))

			newRunT := r.Extract(c1, c2)

			if tk.IsError(newRunT) {
				return p.Errf(r, "failed to runCall: %v", newRunT)
			}

			vs := p.ParamsToList(r, instrT, v1p+2)
			rs := RunCodePiece(NewVMQuick(), newRunT, "", vs, true)

			p.SetVar(r, pr, rs)
			return ""

		}

		vs := p.ParamsToList(r, instrT, v1p+1)

		rs := RunCode(codeT, vs, nil)

		p.SetVar(r, pr, rs)

		return ""

	case 1055: // runCall
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		var rT interface{} = nil

		codeT := p.GetVarValue(r, instrT.Params[v1p])

		if r1, ok := codeT.(*RunningContext); ok {
			vs := p.ParamsToList(r, instrT, v1p+1)
			rs := RunCodePiece(p, r1, "", vs, true)

			p.SetVar(r, pr, rs)
			return ""

		}

		if c1, ok := codeT.(int); ok {
			if instrT.ParamLen < 3 {
				return p.Errf(r, "not enough parameters(参数不够)")
			}

			c2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))

			newRunT := r.Extract(c1, c2)

			if tk.IsError(newRunT) {
				return p.Errf(r, "failed to runCall: %v", newRunT)
			}

			vs := p.ParamsToList(r, instrT, v1p+2)
			rs := RunCodePiece(p, newRunT, "", vs, true)

			p.SetVar(r, pr, rs)
			return ""

		}

		vs := p.ParamsToList(r, instrT, v1p+1)

		if s1, ok := codeT.(string); ok {
			compiledT := Compile(s1)

			if tk.IsError(compiledT) {
				p.SetVar(r, pr, compiledT)
				return ""
			}

			codeT = compiledT
		}

		if cp1, ok := codeT.(*CompiledCode); ok {
			rs := RunCodePiece(p, rT, cp1, vs, true)

			p.SetVar(r, pr, rs)
			return ""
		}

		p.SetVar(r, pr, fmt.Errorf("failed to compile code: %v", codeT))
		return ""
	case 1056: // goRunCall
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		var rT interface{} = nil

		codeT := p.GetVarValue(r, instrT.Params[v1p])

		if r1, ok := codeT.(*RunningContext); ok {
			vs := p.ParamsToList(r, instrT, v1p+1)
			go RunCodePiece(p, r1, "", vs, true)

			// p.SetVar(r, pr, rs)
			p.SetVar(r, pr, nil)
			return ""

		}

		if c1, ok := codeT.(int); ok {
			if instrT.ParamLen < 2 {
				return p.Errf(r, "not enough parameters(参数不够)")
			}

			c2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))

			newRunT := r.Extract(c1, c2)

			if tk.IsError(newRunT) {
				p.SetVar(r, pr, p.Errf(r, "failed to goRunCall: %v", newRunT))
				return ""
			}

			vs := p.ParamsToList(r, instrT, v1p+2)
			go RunCodePiece(p, newRunT, "", vs, true)

			p.SetVar(r, pr, nil)
			return ""

		}

		vs := p.ParamsToList(r, instrT, v1p+1)

		if s1, ok := codeT.(string); ok {
			compiledT := Compile(s1)

			if tk.IsError(compiledT) {
				p.SetVar(r, pr, p.Errf(r, "failed to compile code: %v", compiledT))
				return ""
			}

			codeT = compiledT
		}

		if cp1, ok := codeT.(*CompiledCode); ok {
			go RunCodePiece(p, rT, cp1, vs, true)

			p.SetVar(r, pr, nil)
			return ""
		}

		// return p.Errf(r, "failed to goRunCall code: %v", codeT)
		p.SetVar(r, pr, p.Errf(r, "failed to goRunCall code: %v", codeT))
		return ""
	case 1060: // threadCall/goCall
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		codeT := p.GetVarValue(r, instrT.Params[v1p])

		if r1, ok := codeT.(*RunningContext); ok {
			vs := p.ParamsToList(r, instrT, v1p+1)

			rs := GoCall(r1, vs...)

			p.SetVar(r, pr, rs)
			return ""

		}

		if c1, ok := codeT.(int); ok {
			if instrT.ParamLen < 2 {
				return p.Errf(r, "not enough parameters(参数不够)")
			}

			c2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))

			newRunT := r.Extract(c1, c2)

			if tk.IsError(newRunT) {
				return p.Errf(r, "failed to goCall: %v", newRunT)
			}

			// tk.Pln("newRunT:", newRunT)

			vs := p.ParamsToList(r, instrT, v1p+2)

			rs := GoCall(newRunT, vs...)

			p.SetVar(r, pr, rs)
			return ""

		}

		vs := p.ParamsToList(r, instrT, 2)

		rs := GoCall(codeT, vs...)

		p.SetVar(r, pr, rs)

		return ""
	case 1063: // go, no defer here, 避免使用有可能出问题的指令，例如跳转到范围之外的goto等
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1p := 0

		v2 := r.GetLabelIndex(p.GetVarValue(r, instrT.Params[v1p]))

		vs := p.ParamsToList(r, instrT, v1p+1)

		var deleT tk.QuickVarDelegate

		deleT = func(argsA ...interface{}) interface{} {

			tmpPointerT := v2

			for {
				// tk.Pl("RunLine: %v", tmpPointerT)
				rs := RunInstr(p, r, &r.InstrList[tmpPointerT])

				nv, ok := rs.(int)

				if ok {
					tmpPointerT = nv
					continue
				}

				if tk.IsError(rs) {
					return rs
				}

				nsv, ok := rs.(string)

				if ok {
					if tk.IsErrStr(nsv) {
						return nsv
					}

					if nsv == "exit" {
						return nil
					} else if nsv == "fr" {
						return nil
					}
				}

				tmpPointerT++
			}

			return nil
		}

		go deleT(vs...)

		return ""

	case 1070: // fastCall
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0])

		c1 := r.GetLabelIndex(v1)

		if c1 < 0 {
			return p.Errf(r, "invalid label: %v", v1)
		}

		r.PointerStack.Push(CallStruct{Type: 1, ReturnPointer: r.CodePointer})

		return c1

	case 1071: // fastRet
		rs := r.PointerStack.Pop()

		if tk.IsUndefined(rs) {
			return p.Errf(r, "pointer stack empty")
		}

		nv, ok := rs.(CallStruct)

		if !ok {
			return p.Errf(r, "not in a call, not a call struct in running stack: %v", rs)
		}

		if nv.Type != 1 {
			return p.Errf(r, "not in a fastCall: %v", nv)
		}

		return tk.ToInt(nv.ReturnPointer) + 1
	case 1080: // for
		if instrT.ParamLen < 4 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1p := 0
		v0 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v1 := instrT.Params[v1p+1]
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))
		v3 := p.GetVarValue(r, instrT.Params[v1p+3])

		var v4 interface{} = ":+1"

		if instrT.ParamLen > 4 {
			v4 = p.GetVarValue(r, instrT.Params[v1p+4])
		}

		compiled0T := Compile(v0)

		if tk.IsError(compiled0T) {
			return p.Errf(r, "failed to compile initial instr: %v", compiled0T)
		}

		compiledObj0T := compiled0T.(*CompiledCode)

		if len(compiledObj0T.InstrList) > 0 {
			rs1 := RunInstr(p, r, &compiledObj0T.InstrList[0])

			if tk.IsError(rs1) {
				return p.Errf(r, "failed to run initial instr: %v", rs1)
			}
		}

		compiledT := Compile(v2)

		if tk.IsError(compiledT) {
			return p.Errf(r, "failed to compile instr: %v", v2)
		}

		label1 := r.GetLabelIndex(v3)

		if label1 < 0 {
			return p.Errf(r, "failed to get continue index: %v", v3)
		}

		label2 := r.GetLabelIndex(v4)

		if label2 < 0 {
			return p.Errf(r, "failed to get break index: %v", v4)
		}

		rs := EvalCondition(v1, p, r)

		if tk.IsError(rs) {
			return p.Errf(r, "failed to eval condition: %v", v1)
		}

		rsbT := rs.(bool)

		if rsbT {
			var instr1 *Instr = nil
			compiled1 := compiledT.(*CompiledCode)
			if len(compiled1.InstrList) > 0 {
				instr1 = &compiled1.InstrList[0]
			}

			r.PointerStack.Push(LoopStruct{Cond: v1, LoopInstr: instr1, LoopIndex: label1, BreakIndex: label2})

			return label1
		}

		return label2

		return ""

	case 1085: // range
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0])

		v2 := p.GetVarValue(r, instrT.Params[1])

		var v3 interface{} = ":+1"
		if instrT.ParamLen > 2 {
			v3 = p.GetVarValue(r, instrT.Params[2])
		}

		vs := p.ParamsToList(r, instrT, 3)

		label1 := r.GetLabelIndex(v2)

		if label1 < 0 {
			return p.Errf(r, "failed to get continue index: %v", v2)
		}

		label2 := r.GetLabelIndex(v3)

		if label2 < 0 {
			return p.Errf(r, "failed to get break index: %v", v3)
		}

		iteratorT := tk.NewCompactIterator(v1, vs...)

		if iteratorT == nil {
			return p.Errf(r, "failed to create iterator: %v(%v)", v1, instrT.Params[0])
		}
		// tk.Plv(iteratorT)

		if !iteratorT.HasNext() {
			return label2
		}

		r.PointerStack.Push(RangeStruct{Iterator: iteratorT, LoopIndex: label1, BreakIndex: label2})

		return label1
	case 1087: // getIter
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr1 := instrT.Params[0]
		pr2 := instrT.Params[1]

		objT := r.PointerStack.Peek()

		rangerT, ok := objT.(RangeStruct)

		if !ok {
			p.SetVar(r, pr1, fmt.Errorf("not a range struct: %T(%#v)", objT, objT))
			p.SetVar(r, pr2, tk.Undefined)

			return ""
		}

		countT, kiT, valueT, b1 := rangerT.Iterator.Next()

		// tk.Plo(countT, kiT, valueT, b1)

		p.SetVar(r, pr1, kiT)
		p.SetVar(r, pr2, valueT)

		if instrT.ParamLen > 2 {
			p.SetVar(r, instrT.Params[2], countT)
		}

		if instrT.ParamLen > 3 {
			p.SetVar(r, instrT.Params[3], b1)
		}

		return ""
	case 1101: // newList/newArray
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		vs := p.ParamsToList(r, instrT, v1p)

		vr := make([]interface{}, 0, len(vs))

		for _, v := range vs {
			vr = append(vr, v)
		}

		p.SetVar(r, pr, vr)

		return ""

	case 1110: // addItem/addArrayItem/addListItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// varsT := (*(p.CurrentVarsM))

		pr := instrT.Params[0]

		// p1 := instrT.Params[0].Ref
		v1 := p.GetVarValue(r, instrT.Params[0])

		// v1 := *p1

		// tk.Plo(pr, v1)

		// tk.Pln(p1, instrT, p)
		// tk.Pl("[Line: %v] STACK: %v, CONTEXT: %v", p.CodeSourceMapM[p.CodePointerM]+1, p.StackM[:p.StackPointerM], tk.ToJSONX(p.CurrentFuncContextM, "-inddent", "-sort"))

		// v2 := p.GetVarValue(r, instrT.Params[1])

		// varsT := p.GetVarValue(r, instrT.Params[1])

		// tk.Pln(p1, v2, varsT[p1])
		// varsT[p1] = append((varsT[p1]).([]interface{}), v2)

		var v2 interface{}

		// if instrT.ParamLen < 2 {
		// 	v2 = p. //p.Stack.Pop()
		// } else {
		v2 = p.GetVarValue(r, instrT.Params[1])
		// }
		// *p1 = append((*p1).([]interface{}), v2)

		switch nv := v1.(type) {
		case []interface{}:
			v1 = append(nv, v2)
			p.SetVar(r, pr, v1)

		// 	*p1 = append((*p1).([]interface{}), v2)
		case []bool:
			v1 = append(nv, tk.ToBool(v2))
			p.SetVar(r, pr, v1)
		// 	*p1 = append(nv, tk.ToBool(v2))
		case []int:
			v1 = append(nv, tk.ToInt(v2))
			p.SetVar(r, pr, v1)
		// 	*p1 = append(nv, tk.ToInt(v2))
		case []byte:
			v1 = append(nv, tk.ToByte(v2))
			p.SetVar(r, pr, v1)
		// 	*p1 = append(nv, byte(tk.ToInt(v2)))
		case []rune:
			v1 = append(nv, tk.ToRune(v2))
			p.SetVar(r, pr, v1)
		// 	*p1 = append(nv, rune(tk.ToInt(v2)))
		case []int64:
			v1 = append(nv, int64(tk.ToInt(v2)))
			p.SetVar(r, pr, v1)
		// 	*p1 = append(nv, int64(tk.ToInt(v2)))
		case []float64:
			v1 = append(nv, tk.ToFloat(v2))
			p.SetVar(r, pr, v1)
		// 	*p1 = append(nv, tk.ToFloat(v2))
		case []string:
			v1 = append(nv, tk.ToStr(v2))
			p.SetVar(r, pr, v1)
		// 	*p1 = append(nv, tk.ToStr(v2))
		default:
			valueT := reflect.ValueOf(v1)

			kindT := valueT.Kind()

			if kindT == reflect.Array || kindT == reflect.Slice {
				vrs := reflect.Append(valueT, reflect.ValueOf(v2))

				p.SetVar(r, pr, vrs.Interface())
				return ""
			}

			return p.Errf(r, "参数类型错误：%T(%v) -> %T", v1, nv, v2)
			// tk.Pln(p.ErrStrf("参数类型：%T(%v) -> %T", nv, nv, v2))
		}

		return ""

	case 1111: // addStrItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// varsT := p.GetVars()

		// p1 := instrT.Params[0].Ref
		v1 := p.GetVarValue(r, instrT.Params[0])

		v2 := p.GetVarValue(r, instrT.Params[1])

		// varsT[p1] = append((varsT[p1]).([]string), tk.ToStr(v2))
		v1 = append(v1.([]string), tk.ToStr(v2))

		return ""

	case 1112: // deleteItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// varsT := (*(p.CurrentVarsM))

		// p1 := instrT.Params[0].Ref

		// p1 := p.GetVarValue(r, instrT.Params[0]) // instrT.Params[0].Ref
		// p1 := p.GetVarRef(instrT.Params[0])
		pr := instrT.Params[0]
		v1 := p.GetVarValue(r, instrT.Params[0])

		v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[1]))

		// varsT := p.GetVars()

		// aryT := (varsT[p1]).([]interface{})
		// v1 := *p1

		// aryT := (*p1).([]interface{})

		// if v2 >= len(aryT) {
		// 	return p.Errf(r, "序号超出范围：%v/%v", v2, len(aryT))
		// }

		// rs := make([]interface{}, 0, len(aryT)-1)
		// rs = append(rs, aryT[:v2]...)
		// rs = append(rs, aryT[v2+1:]...)

		// (*p1) = rs

		switch nv := v1.(type) {
		case []interface{}:
			lenT := len(nv)

			if v2 >= lenT {
				return p.Errf(r, "序号超出范围：%v/%v", v2, lenT)
			}

			rs := append(nv[:v2], nv[v2+1:]...)

			p.SetVar(r, pr, rs)
			return ""

		case []bool:
			lenT := len(nv)

			if v2 >= lenT {
				return p.Errf(r, "序号超出范围：%v/%v", v2, lenT)
			}

			rs := append(nv[:v2], nv[v2+1:]...)

			p.SetVar(r, pr, rs)
			return ""

		case []int:
			lenT := len(nv)

			if v2 >= lenT {
				return p.Errf(r, "序号超出范围：%v/%v", v2, lenT)
			}

			rs := append(nv[:v2], nv[v2+1:]...)

			p.SetVar(r, pr, rs)
			return ""

		case []byte:
			lenT := len(nv)

			if v2 >= lenT {
				return p.Errf(r, "序号超出范围：%v/%v", v2, lenT)
			}

			rs := append(nv[:v2], nv[v2+1:]...)

			p.SetVar(r, pr, rs)
			return ""

		case []rune:
			lenT := len(nv)

			if v2 >= lenT {
				return p.Errf(r, "序号超出范围：%v/%v", v2, lenT)
			}

			rs := append(nv[:v2], nv[v2+1:]...)

			p.SetVar(r, pr, rs)
			return ""

		case []int64:
			lenT := len(nv)

			if v2 >= lenT {
				return p.Errf(r, "序号超出范围：%v/%v", v2, lenT)
			}

			rs := append(nv[:v2], nv[v2+1:]...)

			p.SetVar(r, pr, rs)
			return ""

		case []float64:
			lenT := len(nv)

			if v2 >= lenT {
				return p.Errf(r, "序号超出范围：%v/%v", v2, lenT)
			}

			rs := append(nv[:v2], nv[v2+1:]...)

			p.SetVar(r, pr, rs)
			return ""

		case []string:
			lenT := len(nv)

			if v2 >= lenT {
				return p.Errf(r, "序号超出范围：%v/%v", v2, lenT)
			}

			rs := append(nv[:v2], nv[v2+1:]...)

			p.SetVar(r, pr, rs)
			return ""

		default:
			valueT := reflect.ValueOf(v1)

			kindT := valueT.Kind()

			if kindT == reflect.Array || kindT == reflect.Slice {
				// vrs := reflect.Append(valueT, reflect.ValueOf(v2))
				lenT := valueT.Len()

				if v2 >= lenT {
					return p.Errf(r, "序号超出范围：%v/%v", v2, lenT)
				}

				// rs := make([]interface{}, 0, lenT-1)

				// rs = append(rs, aryT[:v2]...)
				// rs = append(rs, aryT[v2+1:]...)

				vrs := reflect.AppendSlice(valueT.Slice(0, v2), valueT.Slice(v2+1, lenT))

				p.SetVar(r, pr, vrs.Interface())
				return ""
			}

			return p.Errf(r, "参数类型错误：%T(%v) -> %T", v1, nv, v2)
		}

		return ""

	case 1115: // addItems
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// varsT := (*(p.CurrentVarsM))

		pr := instrT.Params[0]
		v1p := 0

		if instrT.ParamLen > 2 {
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		switch nv := v1.(type) {
		case []interface{}:
			p.SetVar(r, pr, append(nv, (v2.([]interface{}))...))
		case []bool:
			p.SetVar(r, pr, append(nv, (v2.([]bool))...))
		case []int:
			p.SetVar(r, pr, append(nv, (v2.([]int))...))
		case []byte:
			p.SetVar(r, pr, append(nv, (v2.([]byte))...))
		case []rune:
			p.SetVar(r, pr, append(nv, (v2.([]rune))...))
		case []int64:
			p.SetVar(r, pr, append(nv, (v2.([]int64))...))
		case []float64:
			p.SetVar(r, pr, append(nv, (v2.([]float64))...))
		case []string:
			p.SetVar(r, pr, append(nv, (v2.([]string))...))
		default:
			valueT := reflect.ValueOf(v1)
			value2T := reflect.ValueOf(v2)

			kindT := valueT.Kind()
			kind2T := value2T.Kind()

			if !(kind2T == reflect.Array || kind2T == reflect.Slice) {
				return p.Errf(r, "参数类型错误：%T(%v) -> %T(%v)", v1, v1, v2, v2)
			}

			if kindT == reflect.Array || kindT == reflect.Slice {
				vrs := reflect.AppendSlice(valueT, reflect.ValueOf(v2))
				if !vrs.IsValid() {
					return p.Errf(r, "操作失败，类型错误：%T(%v) -> %T(%v)", v1, v1, v2, v2)
				}
				p.SetVar(r, pr, vrs)
				return
			}

			return p.Errf(r, "参数类型错误：%T(%v) -> %T(%v)", v1, v1, v2, v2)
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

	case 1123: // getArrayItem/[]
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr = instrT.Params[0]

		v1 := p.GetVarValue(r, instrT.Params[1])

		if v1 == nil {
			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
				return ""
			} else {
				return p.Errf(r, "object is nil: (%T)%v", v1, v1)
			}

		}

		v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[2]))

		// tk.Pl("v1: %#v, v2: %#v, instr: %#v", v1, v2, instrT)

		switch nv := v1.(type) {
		case []interface{}:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
					return ""
				} else {
					return p.Errf(r, "index out of range: %v/%v", v2, len(nv))
				}
			}
			// tk.Pl("r: %v", nv[v2])
			p.SetVar(r, pr, nv[v2])
		case []bool:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
					return ""
				} else {
					return p.Errf(r, "index out of range: %v/%v", v2, len(nv))
				}
			}
			p.SetVar(r, pr, nv[v2])
		case []int:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
					return ""
				} else {
					return p.Errf(r, "index out of range: %v/%v", v2, len(nv))
				}
			}
			p.SetVar(r, pr, nv[v2])
		case []byte:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
					return ""
				} else {
					return p.Errf(r, "index out of range: %v/%v", v2, len(nv))
				}
			}
			p.SetVar(r, pr, nv[v2])
		case []rune:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
					return ""
				} else {
					return p.Errf(r, "index out of range: %v/%v", v2, len(nv))
				}
			}
			p.SetVar(r, pr, nv[v2])
		case []int64:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
					return ""
				} else {
					return p.Errf(r, "index out of range: %v/%v", v2, len(nv))
				}
			}
			p.SetVar(r, pr, nv[v2])
		case []float64:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
					return ""
				} else {
					return p.Errf(r, "index out of range: %v/%v", v2, len(nv))
				}
			}
			p.SetVar(r, pr, nv[v2])
		case []string:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
					return ""
				} else {
					return p.Errf(r, "index out of range: %v/%v", v2, len(nv))
				}
			}
			p.SetVar(r, pr, nv[v2])
		case []map[string]string:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
					return ""
				} else {
					return p.Errf(r, "index out of range: %v/%v", v2, len(nv))
				}
			}

			p.SetVar(r, pr, nv[v2])
		case []map[string]interface{}:
			if (v2 < 0) || (v2 >= len(nv)) {
				if instrT.ParamLen > 3 {
					p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
					return ""
				} else {
					return p.Errf(r, "index out of range: %v/%v", v2, len(nv))
				}
			}

			p.SetVar(r, pr, nv[v2])
		default:
			valueT := reflect.ValueOf(v1)

			kindT := valueT.Kind()

			if kindT == reflect.Array || kindT == reflect.Slice || kindT == reflect.String {
				lenT := valueT.Len()

				if (v2 < 0) || (v2 >= lenT) {
					return p.Errf(r, "index out of range: %v/%v", v2, lenT)
				}

				p.SetVar(r, pr, valueT.Index(v2).Interface())
				return ""
			}

			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[3]))
			} else {
				p.SetVar(r, pr, tk.Undefined)
			}

			return p.Errf(r, "parameter types not match: %#v", v1)
		}

		return ""

	case 1124: // setItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0])

		v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[1]))

		var v3 interface{}

		if instrT.ParamLen < 3 {
			v3 = p.Stack.Pop()
			// v1.([]interface{})[v2] = p.Stack.Pop()
		} else {
			v3 = p.GetVarValue(r, instrT.Params[2])
			// v1.([]interface{})[v2] = p.GetVarValue(r, instrT.Params[2])
		}

		switch nv := v1.(type) {
		case []interface{}:
			if v2 >= len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = v3
		case []bool:
			if v2 >= len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = tk.ToBool(v3)
		case []int:
			if v2 >= len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = tk.ToInt(v3)
		case []byte:
			if v2 >= len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = byte(tk.ToInt(v3))
		case []rune:
			if v2 >= len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = rune(tk.ToInt(v3))
		case []int64:
			if v2 >= len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = int64(tk.ToInt(v3))
		case []float64:
			if v2 >= len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = tk.ToFloat(v3)
		case []string:
			if v2 >= len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			nv[v2] = tk.ToStr(v3)
		default:
			valueT := reflect.ValueOf(v1)

			kindT := valueT.Kind()

			if kindT == reflect.Array || kindT == reflect.Slice {
				lenT := valueT.Len()

				if v2 < 0 || v2 >= lenT {
					return p.Errf(r, "序号超出范围：%v/%v", v2, lenT)
				}

				valueT.Index(v2).Set(reflect.ValueOf(v3))
				return ""
			}

			return p.Errf(r, "参数类型错误：%T(%#v) -> %T(%#v)", v1, v1, v3, v3)

		}

		return ""

	case 1130: // slice
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2vT := p.GetVarValue(r, instrT.Params[v1p+1])
		v3vT := p.GetVarValue(r, instrT.Params[v1p+2])

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
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, nv[v2:v3])
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
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, nv[v2:v3])
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
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, nv[v2:v3])
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
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, nv[v2:v3])
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
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, nv[v2:v3])
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
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, nv[v2:v3])
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
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, nv[v2:v3])
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
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, nv[v2:v3])
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
				return p.Errf(r, "序号超出范围：%v/%v", v2, len(nv))
			}

			if v3 > len(nv) {
				return p.Errf(r, "序号超出范围：%v/%v", v3, len(nv))
			}

			if instrT.ParamLen > 3 {
				p.SetVar(r, pr, nv[v2:v3])
			} else {
				v1 = nv[v2:v3]
			}
		default:
			valueT := reflect.ValueOf(v1)

			kindT := valueT.Kind()

			if kindT == reflect.Array || kindT == reflect.Slice || kindT == reflect.String {
				lenT := valueT.Len()

				if v2 == -1 {
					if tk.ToStr(v2vT) == "-" {
						v2 = 0
					}
				}

				if v3 < 0 {
					if tk.ToStr(v3vT) == "-" {
						v3 = lenT
					}
				}

				if (v2 < 0) || (v2 > lenT) {
					return p.Errf(r, "index out of range: %v/%v", v2, lenT)
				}

				if (v3 < 0) || (v3 > lenT) {
					return p.Errf(r, "index out of range: %v/%v", v3, lenT)
				}

				p.SetVar(r, pr, valueT.Slice(v2, v3).Interface())
				return ""
			}

			return p.Errf(r, "参数类型错误：%T(%v) -> %T", v1, nv, v2)
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

	case 1210: // continue
		// tk.Plo("continue PointerStack", r.PointerStack)
		levelT := 1
		if instrT.ParamLen > 0 {
			levelT = tk.ToInt(p.GetVarValue(r, instrT.Params[0]), 1)
		}

		for kk := 1; kk < levelT; kk++ {
			r.PointerStack.Pop()
		}

		v1 := r.PointerStack.Peek()

		if tk.IsUndefined(v1) {
			return p.Errf(r, "no loop/range object in pointer stack: %#v", v1)
		}

		switch nv := v1.(type) {
		case LoopStruct:
			if nv.LoopInstr != nil {
				rsc := RunInstr(p, r, nv.LoopInstr)

				if tk.IsError(rsc) {
					return p.Errf(r, "failed to run loop instr(%#v): %v", nv.LoopInstr, rsc)
				}
			}

			rs := EvalCondition(nv.Cond, p, r)

			if tk.IsError(rs) {
				return p.Errf(r, "failed to eval condition: %#v", nv.Cond)
			}

			rsbT := rs.(bool)

			if rsbT {
				return nv.LoopIndex
			} else {
				r.PointerStack.Pop()

				return nv.BreakIndex
			}
		case RangeStruct:
			rsbT := nv.Iterator.HasNext()

			if rsbT {
				return nv.LoopIndex
			} else {
				r.PointerStack.Pop()

				return nv.BreakIndex
			}
		default:
			return p.Errf(r, "unsupport loop/range structure type: %#v", v1)
		}

		return ""

	case 1211: // break
		// tk.Plo("break PointerStack", r.PointerStack)
		levelT := 1
		if instrT.ParamLen > 0 {
			levelT = tk.ToInt(p.GetVarValue(r, instrT.Params[0]), 1)
		}

		for kk := 1; kk < levelT; kk++ {
			r.PointerStack.Pop()
		}

		v1 := r.PointerStack.Pop()

		if tk.IsUndefined(v1) {
			return p.Errf(r, "no loop/range object in pointer stack: %#v", v1)
		}

		switch nv := v1.(type) {
		case LoopStruct:
			r.PointerStack.Pop()

			return nv.BreakIndex
		case RangeStruct:
			r.PointerStack.Pop()

			return nv.BreakIndex
		default:
			return p.Errf(r, "unsupport loop/range structure type: %#v", v1)
		}

		return ""

	case 1310: // setMapItem
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// p1 := instrT.Params[0].Ref

		v1 := p.GetVarValue(r, instrT.Params[0])

		v2o := p.GetVarValue(r, instrT.Params[1])

		v2 := tk.ToStr(v2o)

		var v3 interface{}

		v3 = p.GetVarValue(r, instrT.Params[2])

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
		case url.Values:
			nv.Set(v2, tk.ToStr(v3))
		case *url.Values:
			nv.Set(v2, tk.ToStr(v3))
		default:
			valueT := reflect.ValueOf(v1)

			kindT := valueT.Kind()

			if kindT == reflect.Map {
				valueT.SetMapIndex(reflect.ValueOf(v2o), reflect.ValueOf(v3))
				return ""
			}

			return p.Errf(r, "参数类型错误: %T(%#v)", v1, v1)
		}

		return ""

	case 1312: // deleteMapItem
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// varsT := (*(p.CurrentVarsM))

		// p1 := instrT.Params[0].Ref

		// p1 := p.GetVarValue(r, instrT.Params[0]) // instrT.Params[0].Ref
		// p1 := p.GetVarRef(instrT.Params[0])
		v1 := p.GetVarValue(r, instrT.Params[0])

		v2o := p.GetVarValue(r, instrT.Params[1])

		v2 := tk.ToStr(v2o)

		// varsT := p.GetVars()

		// aryT := (varsT[p1]).([]interface{})
		// mapT := (*p1).(map[string]interface{})

		// v1 := *p1

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
			valueT := reflect.ValueOf(v1)

			kindT := valueT.Kind()

			if kindT == reflect.Map {
				valueT.SetMapIndex(reflect.ValueOf(v2o), reflect.Zero(reflect.TypeOf(v1)))
				return ""
			}

			return p.Errf(r, "参数类型错误: %T(%#v)", v1)
		}

		// rs := make([]interface{}, 0, len(aryT)-1)
		// rs = append(rs, aryT[:v2]...)
		// rs = append(rs, aryT[v2+1:]...)

		// varsT[p1] = rs // append((varsT[p1]).([]interface{}), v2)
		// delete(mapT, v2)

		// tk.DeleteItemInArray(p1.([]interface{}), v2)

		return ""

	case 1320: // getMapItem/{}
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0] // -5
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2o := p.GetVarValue(r, instrT.Params[v1p+1])
		v2 := tk.ToStr(v2o)

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
		case map[string]map[string]string:
			rv, ok = nv[v2]
		case map[string]map[string]interface{}:
			rv, ok = nv[v2]
		default:
			// tk.Plo("here1", v1, v2o)
			valueT := reflect.ValueOf(v1)

			kindT := valueT.Kind()

			if kindT == reflect.Map {
				rv := valueT.MapIndex(reflect.ValueOf(v2o))

				if !rv.IsValid() {
					if instrT.ParamLen > 3 {
						p.SetVar(r, pr, p.GetVarValue(r, instrT.Params[v1p+2]))

					} else {
						p.SetVar(r, pr, tk.Undefined)
					}

					return ""
				}

				p.SetVar(r, pr, rv.Interface())

				return ""
			}

			rv := tk.ReflectGetMember(v1, v2)

			if tk.IsError(rv) {
				if instrT.ParamLen > 3 {
					rv = p.GetVarValue(r, instrT.Params[v1p+2])

				} else {
					rv = tk.Undefined
				}

				p.SetVar(r, pr, rv)
				return ""
			}

			p.SetVar(r, pr, rv)
			return ""
			// return p.Errf(r, "参数类型错误：%T（%v）", v1, v1)
		}

		if !ok {
			if instrT.ParamLen > 3 {
				rv = p.GetVarValue(r, instrT.Params[v1p+2])

			} else {
				rv = tk.Undefined
			}
		}

		p.SetVar(r, pr, rv)

		return ""
	case 1331: // getMapKeys
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		var rv []string

		switch nv := v1.(type) {
		case map[string]interface{}:
			rv = make([]string, 0, len(nv))
			for k, _ := range nv {
				rv = append(rv, k)
			}
		case map[string]int:
			rv = make([]string, 0, len(nv))
			for k, _ := range nv {
				rv = append(rv, k)
			}
		case map[string]byte:
			rv = make([]string, 0, len(nv))
			for k, _ := range nv {
				rv = append(rv, k)
			}
		case map[string]rune:
			rv = make([]string, 0, len(nv))
			for k, _ := range nv {
				rv = append(rv, k)
			}
		case map[string]float64:
			rv = make([]string, 0, len(nv))
			for k, _ := range nv {
				rv = append(rv, k)
			}
		case map[string]string:
			rv = make([]string, 0, len(nv))
			for k, _ := range nv {
				rv = append(rv, k)
			}
		case map[string]map[string]string:
			rv = make([]string, 0, len(nv))
			for k, _ := range nv {
				rv = append(rv, k)
			}
		case map[string]map[string]interface{}:
			rv = make([]string, 0, len(nv))
			for k, _ := range nv {
				rv = append(rv, k)
			}
		default:
			valueT := reflect.ValueOf(v1)

			kindT := valueT.Kind()

			if kindT == reflect.Map {
				keysT := valueT.MapKeys()

				rvo := reflect.MakeSlice(valueT.Type().Key(), 0, len(keysT))

				for _, v := range keysT {
					reflect.Append(rvo, v)
				}

				rvi := rvo.Interface()

				return rvi
			}

			return p.Errf(r, "参数类型错误：%T(%v)", v1, v1)
		}

		p.SetVar(r, pr, rv)

		return ""

	case 1401: // new
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rs := NewObject(p, r, v1, p.ParamsToList(r, instrT, v1p+1)...)

		p.SetVar(r, pr, rs)
		return ""

	case 1403: // method/mt
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v1 := p.GetVarValue(r, instrT.Params[1])

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[2]))

		v3p := 3

		switch nv := v1.(type) {
		case string:
			mapT := memberMapG["string"]

			funcT, ok := mapT[v2]

			if ok {
				rsT := callGoFunc(funcT, nv, p.ParamsToList(r, instrT, v3p)...)
				p.SetVar(r, pr, rsT)

				return ""
			}
		case *tk.SyncQueue:
			switch v2 {
			case "add", "put":
				v3 := p.ParamsToStrs(r, instrT, v3p)

				nv.Add(v3)
				return ""
			case "clearAdd":
				v3 := p.ParamsToStrs(r, instrT, v3p)

				nv.ClearAdd(v3)
				return ""
			case "clear":
				nv.Clear()
				return ""
			case "size":
				p.SetVar(r, pr, nv.Size())
				return ""
			case "quickGet":
				p.SetVar(r, pr, nv.QuickGet())
				return ""
			case "get":
				rs, ok := nv.Get()

				if !ok {
					p.SetVar(r, pr, fmt.Errorf("队列空（queue empty）"))
					return ""
				}

				p.SetVar(r, pr, rs)
				return ""
			}
		case *mailyak.MailYak:
			switch v2 {
			case "to":
				vs := p.ParamsToStrs(r, instrT, 3)

				nv.To(vs...)

				// p.SetVar(r, pr, nil)

				return ""
			case "cc":
				vs := p.ParamsToStrs(r, instrT, 3)

				nv.Cc(vs...)

				return ""
			case "bcc":
				vs := p.ParamsToStrs(r, instrT, 3)

				nv.Bcc(vs...)

				return ""
			case "from":
				if instrT.ParamLen < 4 {
					return p.Errf(r, "not enough parameters(参数不够)")
				}

				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))

				nv.From(v3)

				return ""
			case "fromName":
				if instrT.ParamLen < 4 {
					return p.Errf(r, "not enough parameters(参数不够)")
				}

				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))

				nv.FromName(v3)

				return ""
			case "subject":
				if instrT.ParamLen < 4 {
					return p.Errf(r, "参数不够（not enough parameters）")
				}

				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))

				nv.Subject(v3)

				return ""
			case "replyTo":
				if instrT.ParamLen < 4 {
					return p.Errf(r, "参数不够（not enough parameters）")
				}

				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))

				nv.ReplyTo(v3)

				return ""
			case "writeBccHeader":
				if instrT.ParamLen < 4 {
					return p.Errf(r, "参数不够（not enough parameters）")
				}

				v3 := tk.ToBool(p.GetVarValue(r, instrT.Params[3]))

				nv.WriteBccHeader(v3)

				return ""
			case "clearAttachments":
				nv.ClearAttachments()

				return ""
			case "attach": // 添加附件(最后一个参数是mime类型，可以省略)，用法：mt $drop $mail attach "imageName1.png" $fr1 "image/png"
				if instrT.ParamLen < 5 {
					return p.Errf(r, "参数不够（not enough parameters）")
				}

				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))
				v4 := p.GetVarValue(r, instrT.Params[4]).(io.Reader)

				if instrT.ParamLen > 5 {
					v5 := tk.ToStr(p.GetVarValue(r, instrT.Params[5]))

					nv.AttachWithMimeType(v3, v4, v5)

					return ""
				}

				nv.Attach(v3, v4)

				return ""
			case "attachInline": // （此法未经验证有效）添加内嵌附件以便在html中引用，引用方法是：<img src="cid:myFileName"/>，用法：mt $drop $mail attachInline "imageName1.png" $fr1 "image/png"

				if instrT.ParamLen < 5 {
					return p.Errf(r, "参数不够（not enough parameters）")
				}

				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))
				v4 := p.GetVarValue(r, instrT.Params[4]).(io.Reader)

				if instrT.ParamLen > 5 {
					v5 := tk.ToStr(p.GetVarValue(r, instrT.Params[5]))

					nv.AttachInlineWithMimeType(v3, v4, v5)

					return ""
				}

				nv.AttachInline(v3, v4)

				return ""
			case "body", "setHtmlBody", "setBody":
				if instrT.ParamLen < 4 {
					return p.Errf(r, "参数不够（not enough parameters）")
				}

				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))

				_, errT := io.WriteString(nv.HTML(), v3)

				if errT != nil {
					p.SetVar(r, pr, p.Errf(r, "邮件格式解析错误（failed to parse mail body）：%v", errT))
					return ""
				}

				nv.Plain().Set(v3)

				p.SetVar(r, pr, "")
				return ""
			case "setPlainBody":
				if instrT.ParamLen < 4 {
					return p.Errf(r, "参数不够（not enough parameters）")
				}

				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))

				_, errT := io.WriteString(nv.HTML(), v3)

				if errT != nil {
					p.SetVar(r, pr, p.Errf(r, "邮件格式解析错误（failed to parse mail body）：%v", errT))
					return ""
				}

				nv.Plain().Set(v3)

				p.SetVar(r, pr, "")
				return ""
			case "addHeader":
				if instrT.ParamLen < 5 {
					return p.Errf(r, "参数不够（not enough parameters）")
				}

				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))
				v4 := tk.ToStr(p.GetVarValue(r, instrT.Params[4]))

				nv.AddHeader(v3, v4)

				return ""
			case "send":
				errT := nv.Send()

				if errT != nil {
					p.SetVar(r, pr, p.Errf(r, "邮件发送失败（failed to send mail）：%v", errT))
					return ""
				}

				p.SetVar(r, pr, "")
				return ""
			case "info", "string":
				p.SetVar(r, pr, nv.String())
				return ""
			}
		// case *tk.QuickObject:
		// 	switch nv.Type {
		// 	case "mailSender":
		// 		if nv.Value == nil {
		// 			nv.Value = make(map[string]interface{})
		// 		}

		// 		switch v2 {
		// 		case "setHost":
		// 			v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v3p]))

		// 			valueT := nv.Value.(map[string]interface{})

		// 			valueT["Host"] = v3

		// 			p.SetVar(r, pr, nil)

		// 			return ""
		// 		}

		// 	}
		case time.Time:
			switch v2 {
			case "toStr":
				p.SetVar(r, pr, fmt.Sprintf("%v", nv))
				return ""
			case "toTick":
				p.SetVar(r, pr, tk.GetTimeStampMid(nv))
				return ""
			case "getInfo":
				zoneT, offsetT := nv.Zone()

				p.SetVar(r, pr, map[string]interface{}{"Time": nv, "Formal": nv.Format(tk.TimeFormat), "Compact": nv.Format(tk.TimeFormat), "Full": fmt.Sprintf("%v", nv), "Year": nv.Year(), "Month": nv.Month(), "Day": nv.Day(), "Hour": nv.Hour(), "Minute": nv.Minute(), "Second": nv.Second(), "Zone": zoneT, "Offset": offsetT, "UnixNano": nv.UnixNano()})
				return ""
			case "format":
				var v2 string = ""

				if instrT.ParamLen > 3 {
					v2 = tk.ToStr(p.GetVarValue(r, instrT.Params[3]))
				}

				p.SetVar(r, pr, tk.FormatTime(nv, v2))

				return ""
			case "toLocal":
				p.SetVar(r, pr, nv.Local())
				return ""
			case "toGlobal", "toUTC":
				p.SetVar(r, pr, nv.UTC())
				return ""
			case "addDate":
				if instrT.ParamLen < 6 {
					return p.Errf(r, "not enough paramters")
				}

				v1p := 2

				v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))
				v3 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+2]))
				v4 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+3]))

				p.SetVar(r, pr, nv.AddDate(v2, v3, v4))

				return ""
			case "sub":
				if instrT.ParamLen < 4 {
					return p.Errf(r, "not enough paramters")
				}

				v1p := 2

				v2 := tk.ToTime(p.GetVarValue(r, instrT.Params[v1p+1]))

				vvv := nv.Sub(v2.(time.Time))

				p.SetVar(r, pr, int(vvv)/1000000)
				return ""
			default:
				break
				// p.SetVar(r, pr, fmt.Sprintf("未知方法: %v", v2))
				// return p.ErrStrf("未知方法: %v", v2)
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
				break
				// p.SetVar(r, pr, fmt.Sprintf("未知方法: %v", v2))
				// return p.ErrStrf("未知方法: %v", v2)
			}

			// p.SetVar(r, pr, "")
			// return ""
		case *goph.Client:
			switch v2 {
			case "close":
				nv.Close()
			case "run":
				if instrT.ParamLen < 4 {
					return p.Errf(r, "not enough paramters")
				}

				v1p := 2

				v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

				rs, errT := nv.Run(v2)

				if errT != nil {
					p.SetVar(r, pr, errT)
				}

				p.SetVar(r, pr, string(rs))

				return ""
			case "upload":
				if instrT.ParamLen < 5 {
					return p.Errf(r, "not enough paramters")
				}

				v1p := 2

				v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))
				vs := p.ParamsToStrs(r, instrT, v1p+3)

				p.SetVar(r, pr, nv.Upload(v2, v3, vs...))

				return ""
			case "download":
				if instrT.ParamLen < 5 {
					return p.Errf(r, "not enough paramters")
				}

				v1p := 2

				v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))

				p.SetVar(r, pr, nv.Download(v2, v3))

				return ""
			case "getFileContent":
				if instrT.ParamLen < 4 {
					return p.Errf(r, "not enough paramters")
				}

				v1p := 2

				v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

				rs, errT := nv.GetFileContent(v2)

				if errT != nil {
					p.SetVar(r, pr, errT)
				}

				p.SetVar(r, pr, rs)

				return ""
			default:
				break
				// p.SetVar(r, pr, fmt.Sprintf("未知方法: %v", v2))
				// return p.ErrStrf("未知方法: %v", v2)
			}

			// p.SetVar(r, pr, "")
			// return ""
		case *strings.Builder:

			switch v2 {
			case "write", "append", "writeString":
				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))
				c, errT := nv.WriteString(v3)
				if errT != nil {
					p.SetVar(r, pr, errT)
					return ""
				}

				p.SetVar(r, pr, c)
				return ""
			case "len":
				p.SetVar(r, pr, nv.Len())
				return ""
			case "reset":
				nv.Reset()
				return ""
			case "string", "str", "getStr", "getString":
				p.SetVar(r, pr, nv.String())
				return ""
			default:
				break
				// p.SetVar(r, pr, fmt.Sprintf("未知方法: %v", v2))
				// return p.ErrStrf("未知方法: %v", v2)
			}
			break
			// return ""
		case *tk.Seq:
			switch v2 {
			case "get":
				p.SetVar(r, pr, nv.Get())
				return ""
			default:
				break
				// p.SetVar(r, pr, fmt.Sprintf("未知方法: %v", v2))
				// return p.ErrStrf("未知方法: %v", v2)
			}
			break
		case tk.TXDelegate:
			// if instrT.ParamLen < 3 {
			// 	return p.Errf(r, "not enough paramters")
			// }

			v1p := 2

			v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
			vs := p.ParamsToList(r, instrT, v1p+1)

			rs := nv(v2, p, nil, vs...)

			p.SetVar(r, pr, rs)

			return ""
		case *http.Request:
			switch v2 {
			case "saveFormFile":
				if instrT.ParamLen < 6 {
					return p.Errf(r, "not enough paramters")
				}

				v1p := 2

				v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
				v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))
				v4 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+3]))

				argsT := p.ParamsToStrs(r, instrT, v1p+4)

				formFile1, headerT, errT := nv.FormFile(v2)
				if errT != nil {
					p.SetVar(r, pr, fmt.Sprintf("获取上传文件失败：%v", errT))
					return ""
				}

				defer formFile1.Close()
				tk.Pl("file name : %#v", headerT.Filename)

				defaultExtT := p.GetSwitchVarValue(r, argsT, "-defaultExt=", "")

				baseT := tk.RemoveFileExt(filepath.Base(headerT.Filename))
				extT := filepath.Ext(headerT.Filename)

				if extT == "" {
					extT = defaultExtT
				}

				v4 = strings.Replace(v4, "TX_fileName_XT", baseT, -1)
				v4 = strings.Replace(v4, "TX_fileExt_XT", extT, -1)

				destFile1, errT := os.CreateTemp(v3, v4) //"pic*.png")
				if errT != nil {
					p.SetVar(r, pr, fmt.Sprintf("保存上传文件失败：%v", errT))
					return ""
				}

				defer destFile1.Close()

				_, errT = io.Copy(destFile1, formFile1)
				if errT != nil {
					p.SetVar(r, pr, fmt.Sprintf("服务器内部错误：%v", errT))
					return ""
				}

				p.SetVar(r, pr, tk.GetLastComponentOfFilePath(destFile1.Name()))
				return ""

			}
			// return ""
		}

		mapT := memberMapG[""]

		funcT, ok := mapT[v2]

		if ok {
			rsT := callGoFunc(funcT, v1, p.ParamsToList(r, instrT, 3)...)
			p.SetVar(r, pr, rsT)

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
		rvr := tk.ReflectCallMethod(v1, v2, p.ParamsToList(r, instrT, 3)...)

		p.SetVar(r, pr, rvr)

		// p.SetVar(r, pr, fmt.Errorf("未知方法：（%v）%v", v1, v2))

		return ""
	case 1405: // member/mb
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v1 := p.GetVarValue(r, instrT.Params[1])

		// v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[2]))

		vs := p.ParamsToStrs(r, instrT, 2)

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

				// p.SetVar(r, pr, fmt.Sprintf("未知成员: %v", v2))
				// return p.Errf(r, "未知成员: %v", v2)

			case *url.URL:
				switch v2 {
				case "Scheme":
					vr = nv.Scheme
					continue
				}

				// p.SetVar(r, pr, fmt.Sprintf("未知成员: %v", v2))
				// return p.Errf(r, "未知成员: %v", v2)

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

			return p.Errf(r, "未知成员：%v（%T/%v）.%v", vr, vr, kindT, v2)

		}

		p.SetVar(r, pr, vr)
		return ""
	case 1407: // mbSet
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		v3 := p.GetVarValue(r, instrT.Params[v1p+2])

		// var vr interface{} = v1

		switch nv := v1.(type) {
		case string:
			return p.Errf(r, "无法处理的类型：%v（%T/%v）", nv, v1, v1)
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

			return p.Errf(r, "无法处理的类型：%v（%T/%v）", v1, v1, kindT)

		}

		p.SetVar(r, pr, "")
		return ""
	case 1410: // newObj
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))

		switch v1 {
		case "string":
			objT := &XieString{}

			objT.Init(p.ParamsToList(r, instrT, 2)...)

			p.SetVar(r, pr, objT)
		case "any":
			objT := &XieAny{}

			objT.Init(p.ParamsToList(r, instrT, 2)...)

			p.SetVar(r, pr, objT)
		case "mux":
			p.SetVar(r, pr, http.NewServeMux())
		case "int":
			p.SetVar(r, pr, new(int))
		case "byte":
			p.SetVar(r, pr, new(byte))
		default:
			return p.Errf(r, "未知对象")
		}

		return ""
	case 1411: // setObjValue
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0]).(XieObject)

		// v2 := p.GetVarValue(r, instrT.Params[0])

		v1.SetValue(p.ParamsToList(r, instrT, 1)...)

		return ""
	case 1412: // getObjValue
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p]).(XieObject)

		p.SetVar(r, pr, v1.GetValue())

		return ""
	case 1440: // callObj
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p]).(XieObject)

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		argsT := []interface{}{v2}

		argsT = append(argsT, p.ParamsToList(r, instrT, v1p+2)...)

		rsT := v1.Call(argsT...)

		if tk.IsError(rsT) {
			return p.Errf(r, "对象方法调用失败：%v", rsT)
		}

		p.SetVar(r, pr, rsT)

		return ""

	case 1501: // backQuote
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		p.SetVar(r, pr, "`")

		return ""
	case 1503: // quote
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rs := strconv.Quote(v1)

		p.SetVar(r, pr, rs[1:len(rs)-1])

		return ""
	case 1504: // unquote
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rs, errT := strconv.Unquote(`"` + v1 + `"`)

		if errT != nil {
			p.Errf(r, "unquote失败：%v", errT)
		}

		p.SetVar(r, pr, rs)

		return ""
	case 1510: // isEmpty
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, (v1 == ""))

		return ""

	case 1513: // isEmptyTrim
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.Trim(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])))

		p.SetVar(r, pr, (v1 == ""))

		return ""

	case 1520: // strAdd
		var pr interface{} = -5
		var v1, v2 interface{}

		if instrT.ParamLen == 0 {
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
		} else if instrT.ParamLen == 1 {
			pr = instrT.Params[0]
			v2 = p.Stack.Pop()
			v1 = p.Stack.Pop()
			// return p.Errf(r, "not enough parameters(参数不够)")
		} else if instrT.ParamLen == 2 {
			v1 = p.GetVarValue(r, instrT.Params[0])
			v2 = p.GetVarValue(r, instrT.Params[1])
		} else {
			pr = instrT.Params[0]
			v1 = p.GetVarValue(r, instrT.Params[1])
			v2 = p.GetVarValue(r, instrT.Params[2])
		}

		p.SetVar(r, pr, v1.(string)+v2.(string))

		return ""

	case 1530: // strSplit
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		s1 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))

		s2 := tk.ToStr(p.GetVarValue(r, instrT.Params[2]))

		countT := -1

		if instrT.ParamLen > 3 {
			countT = tk.ToInt(p.GetVarValue(r, instrT.Params[3]))
		}

		listT := strings.SplitN(s1, s2, countT)

		p.SetVar(r, pr, listT)

		return ""

	case 1533: // strSplitByLen
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		s1 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))

		c2 := tk.ToInt(p.GetVarValue(r, instrT.Params[2]))

		vs := p.ParamsToList(r, instrT, 3)

		listT := tk.SplitByLen(s1, c2, vs...)

		p.SetVar(r, pr, listT)

		return ""

	case 1540: // strReplace
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))

		// tk.Plo(v1, v2, v3)

		p.SetVar(r, pr, strings.ReplaceAll(v1, v2, v3))

		return ""

	case 1543: // strReplaceIn
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		vs := p.ParamsToStrs(r, instrT, v1p+1)

		p.SetVar(r, pr, tk.StringReplace(v1, vs...))

		return ""
	case 1550: // trim/strTrim
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		if v1 == nil {
			p.SetVar(r, pr, "")
		} else if v1 == tk.Undefined {
			p.SetVar(r, pr, "")
		} else {
			p.SetVar(r, pr, strings.TrimSpace(tk.ToStr(v1)))
		}

		return ""

	case 1551: // trimSet
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		p.SetVar(r, pr, strings.Trim(v1, v2))

		return ""

	case 1553: // trimSetLeft
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		p.SetVar(r, pr, strings.TrimLeft(v1, v2))

		return ""

	case 1554: // trimSetRight
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		p.SetVar(r, pr, strings.TrimRight(v1, v2))

		return ""

	case 1557: // trimPrefix
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		p.SetVar(r, pr, strings.TrimPrefix(v1, v2))

		return ""

	case 1558: // trimSuffix
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		p.SetVar(r, pr, strings.TrimSuffix(v1, v2))

		return ""

	case 1561: // toUpper
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, strings.ToUpper(v1))

		return ""

	case 1562: // toLower
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, strings.ToLower(v1))

		return ""

	case 1563: // strPad
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))
		v3 := tk.ToInt(p.GetVarValue(r, instrT.Params[2]))

		p.SetVar(r, p1, tk.PadString(v2, v3, p.ParamsToStrs(r, instrT, 2)...))

		return ""

	case 1571: // strContains
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		p.SetVar(r, pr, strings.Contains(v1, v2))

		return ""

	case 1572: // strContainsIn
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := p.ParamsToStrs(r, instrT, v1p+1)

		p.SetVar(r, pr, tk.ContainsIn(v1, v2...))

		return ""

	case 1573: // strCount
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		p.SetVar(r, pr, strings.Count(v1, v2))

		return ""

	case 1581: // strIn/inStrs
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))

		v3i, ok := p.GetVarValue(r, instrT.Params[2]).([]string)

		if ok {
			p.SetVar(r, pr, tk.InStrings(v2, v3i...))
			return ""
		}

		v3 := p.ParamsToStrs(r, instrT, 2)

		p.SetVar(r, pr, tk.InStrings(v2, v3...))

		return ""

	case 1582: // strStartsWith
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		p.SetVar(r, pr, strings.HasPrefix(v1, v2))

		return ""

	case 1583: // strStartsWithIn
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := p.ParamsToStrs(r, instrT, v1p+1)

		p.SetVar(r, pr, tk.StartsWith(v1, v2...))

		return ""

	case 1584: // strEndsWith
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		p.SetVar(r, pr, strings.HasSuffix(v1, v2))

		return ""

	case 1585: // strEndsWithIn
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := p.ParamsToStrs(r, instrT, v1p+1)

		p.SetVar(r, pr, tk.EndsWith(v1, v2...))

		return ""
	case 1601: // bytesToData
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p]).([]byte)
		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		vs := p.ParamsToStrs(r, instrT, v1p+2)

		p.SetVar(r, pr, tk.BytesToData(v1, v2, vs...))

		return ""

	case 1603: // dataToBytes
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		vs := p.ParamsToStrs(r, instrT, v1p+1)

		p.SetVar(r, pr, tk.DataToBytes(v1, vs...))

		return ""

	case 1605: // bytesToHex
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p]).([]byte)

		vs := p.ParamsToStrs(r, instrT, v1p+1)

		p.SetVar(r, pr, tk.BytesToHex(v1, vs...))

		return ""

	case 1606: // bytesToHexX
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p]).([]byte)

		p.SetVar(r, pr, tk.BytesToHexX(v1))

		return ""

	case 1701: // lock
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// pr := instrT.Params[0]
		v1p := 0

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		switch nv := v1.(type) {
		case *sync.RWMutex:
			nv.Lock()

			return ""
		case *sync.Mutex:
			nv.Lock()

			return ""
		case sync.Mutex:
			nv.Lock()

			return ""

		default:
			return p.Errf(r, "不支持的类型(type not supported)：%T(%v)", v1, v1)

		}

		// tk.Pl("v1: %T(%#v), nv1: %#v, ok: %v", v1, v1, nv1, ok)

		return ""

	case 1703: // unlock
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// pr := instrT.Params[0]
		v1p := 0

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		switch nv := v1.(type) {
		case *sync.RWMutex:
			nv.Unlock()

			return ""
		case *sync.Mutex:
			nv.Unlock()

			return ""
		case sync.Mutex:
			nv.Unlock()

			return ""

		default:
			return p.Errf(r, "不支持的类型(type not supported)：%T(%v)", v1, v1)

		}

		return ""

	case 1721: // lockN
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// pr := instrT.Params[0]
		v1p := 0

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]), 0)

		tk.LockN(v1)

		return ""

	case 1722: // unlockN
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// pr := instrT.Params[0]
		v1p := 0

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]), 0)

		tk.UnlockN(v1)

		return ""

	case 1723: // tryLockN
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]), 0)

		p.SetVar(r, pr, tk.TryLockN(v1))

		return ""

	case 1725: // readLockN
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// pr := instrT.Params[0]
		v1p := 0

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]), 0)

		tk.RLockN(v1)

		return ""

	case 1726: // readUnlockN
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// pr := instrT.Params[0]
		v1p := 0

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]), 0)

		tk.RUnlockN(v1)

		return ""

	case 1727: // tryReadLockN
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]), 0)

		p.SetVar(r, pr, tk.TryRLockN(v1))

		return ""

	case 1910: // now

		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		errT := p.SetVar(r, pr, time.Now())

		if errT != nil {
			return p.Errf(r, "%v", errT)
		}

		return ""

	case 1911: // nowStrCompact
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		errT := p.SetVar(r, pr, tk.GetNowTimeString())

		if errT != nil {
			return p.Errf(r, "%v", errT)
		}

		return ""

	case 1912: // nowStr/nowStrFormal
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		errT := p.SetVar(r, pr, tk.GetNowTimeStringFormal())

		if errT != nil {
			return p.Errf(r, "%v", errT)
		}

		return ""
	case 1913: // nowTick
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		errT := p.SetVar(r, pr, tk.GetTimeStampMid(time.Now()))

		if errT != nil {
			return p.Errf(r, "%v", errT)
		}

		return ""
	case 1918: // nowUTC
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		errT := p.SetVar(r, pr, time.Now().UTC())

		if errT != nil {
			return p.Errf(r, "%v", errT)
		}

		return ""

	case 1921: // timeSub
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		sd := int(v1.(time.Time).Sub(v2.(time.Time)) / time.Millisecond)

		p.SetVar(r, pr, sd)

		return ""

	case 1941: // timeToLocal
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToTime(p.GetVarValue(r, instrT.Params[v1p]))

		if tk.IsError(v1) {
			return p.Errf(r, "时间转换失败：%v", v1)
		}

		p.SetVar(r, pr, v1.(time.Time).Local())

		return ""

	case 1942: // timeToGlobal
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToTime(p.GetVarValue(r, instrT.Params[v1p]))

		if tk.IsError(v1) {
			return p.Errf(r, "时间转换失败：%v", v1)
		}

		p.SetVar(r, pr, v1.(time.Time).UTC())

		return ""

	case 1951: // getTimeInfo
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToTime(p.GetVarValue(r, instrT.Params[v1p]))

		if tk.IsError(v1) {
			return p.Errf(r, "时间转换失败：%v", v1)
		}

		nv := v1.(time.Time)

		zoneT, offsetT := nv.Zone()

		p.SetVar(r, pr, map[string]interface{}{"Time": nv, "Formal": nv.Format(tk.TimeFormat), "Compact": nv.Format(tk.TimeFormat), "Full": fmt.Sprintf("%v", nv), "Year": nv.Year(), "Month": nv.Month(), "Day": nv.Day(), "Hour": nv.Hour(), "Minute": nv.Minute(), "Second": nv.Second(), "Zone": zoneT, "Offset": offsetT})

		return ""

	case 1961: // timeAddDate
		if instrT.ParamLen < 5 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToTime(p.GetVarValue(r, instrT.Params[v1p]))

		if tk.IsError(v1) {
			return p.Errf(r, "时间转换失败：%v", v1)
		}

		v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))
		v3 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+2]))
		v4 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+3]))

		nv := v1.(time.Time)

		p.SetVar(r, pr, nv.AddDate(v2, v3, v4))

		return ""

	case 1971: // formatTime
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1o := p.GetVarValue(r, instrT.Params[v1p])

		v1 := tk.ToTime(v1o)

		if tk.IsError(v1) {
			return p.Errf(r, "failed to convert time(时间转换失败): %v(%#v)", v1, v1o)
		}

		var v2 string = ""

		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
		}

		p.SetVar(r, pr, tk.FormatTime(v1.(time.Time), v2))

		return ""

	case 1991: // timeToTick
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToTime(p.GetVarValue(r, instrT.Params[v1p]), p.ParamsToList(r, instrT, v1p+1)...)

		if tk.IsErrX(v1) {
			p.SetVar(r, pr, v1)
			return ""
		}

		t := tk.GetTimeStampMid(v1.(time.Time))

		p.SetVar(r, pr, t)

		return ""

	case 1993: // tickToTime
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]))

		if v1 == -1 {
			p.SetVar(r, pr, fmt.Errorf("转换时间戳失败：%v", p.GetVarValue(r, instrT.Params[v1p])))
			return ""
		}

		t := time.Unix(int64(v1), 0)

		p.SetVar(r, pr, t)

		return ""

	case 2100: // abs
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough paramters")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, tk.Abs(v1))

		return ""

	case 10001: // getParam
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough paramters")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := p.GetVarValue(r, instrT.Params[v1p+2])

		v1n, ok := v1.([]string)

		if ok {
			p.SetVar(r, pr, tk.GetParameterByIndexWithDefaultValue(v1n, tk.ToInt(v2), tk.ToStr(v3)))
			return ""
		}

		v2n, ok := v1.([]interface{})

		if ok {
			p.SetVar(r, pr, tk.GetParamI(v2n, tk.ToInt(v2), tk.ToStr(v3)))
			return ""
		}

		return p.Errf(r, "invalid parameter type: %T(%#v)", v1)

	case 10002: // getSwitch
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough paramters")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			v1p = 1
			pr = instrT.Params[0]
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		defaultT := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))

		v1n, ok := v1.([]string)

		if ok {
			p.SetVar(r, pr, tk.GetSwitch(v1n, v2, defaultT))
			return ""
		}

		v2n, ok := v1.([]interface{})

		if ok {
			p.SetVar(r, pr, tk.GetSwitchI(v2n, v2, defaultT))
			return ""
		}

		return p.Errf(r, "invalid parameter type: %T(%#v)", v1)

	case 10003: // ifSwitchExists
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough paramters")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			v1p = 1
			pr = instrT.Params[0]
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v1n, ok := v1.([]string)

		if ok {
			p.SetVar(r, pr, tk.IfSwitchExistsWhole(v1n, tk.ToStr(v2)))
			return ""
		}

		v2n, ok := v1.([]interface{})

		if ok {
			p.SetVar(r, pr, tk.IfSwitchExistsWholeI(v2n, tk.ToStr(v2)))
			return ""
		}

		return p.Errf(r, "invalid parameter type: %T(%#v)", v1)

	case 10005: // ifSwitchNotExists/switchNotExists
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough paramters")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			v1p = 1
			pr = instrT.Params[0]
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v1n, ok := v1.([]string)

		if ok {
			p.SetVar(r, pr, !tk.IfSwitchExistsWhole(v1n, tk.ToStr(v2)))
			return ""
		}

		v2n, ok := v1.([]interface{})

		if ok {
			p.SetVar(r, pr, !tk.IfSwitchExistsWholeI(v2n, tk.ToStr(v2)))
			return ""
		}

		return p.Errf(r, "invalid parameter type: %T(%#v)", v1)
	case 10011: // parseCommandLine
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough paramters")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		rs, errT := tk.ParseCommandLine(tk.ToStr(v1))
		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		p.SetVar(r, pr, rs)
		return ""

	case 10410: // pln
		list1T := []interface{}{}

		for _, v := range instrT.Params {
			list1T = append(list1T, p.GetVarValue(r, v))
		}

		fmt.Println(list1T...)

		return ""

	case 10411: // plo
		if instrT.ParamLen < 1 {
			tk.Plo(p.GetCurrentFuncContext(r).Tmp)
			return ""
		}

		vs := p.ParamsToList(r, instrT, 0)

		tk.Plo(vs...)

		return ""
	case 10412: // plos
		if instrT.ParamLen < 1 {
			tk.Plos(p.GetCurrentFuncContext(r).Tmp)
			return ""
		}

		vs := p.ParamsToList(r, instrT, 0)

		tk.Plos(vs...)

		return ""
	case 10415: // pr
		list1T := []interface{}{}

		for _, v := range instrT.Params {
			list1T = append(list1T, p.GetVarValue(r, v))
		}

		tk.Pr(list1T...)

		return ""

	case 10416: // prf
		list1T := []interface{}{}

		formatT := ""

		for i, v := range instrT.Params {
			if i == 0 {
				formatT = tk.ToStr(v.Value)
				continue
			}

			list1T = append(list1T, p.GetVarValue(r, v))
		}

		tk.Prf(formatT, list1T...)

		return ""

	case 10420: // pl
		list1T := []interface{}{}

		formatT := ""

		// tk.Pl("instrT.Params: %v", instrT.Params)

		for i, v := range instrT.Params {
			// tk.Pl("[%v]: %v %#v", i, v, v)
			if i == 0 {
				formatT = tk.ToStr(p.GetVarValue(r, v))
				continue
			}

			if v.Ref != -3 && v.Value == "..." {
				vv := p.GetVarValue(r, v)
				// tk.Plo(vv)

				switch nv := vv.(type) {
				case []byte:
					for _, v9 := range nv {
						list1T = append(list1T, v9)
					}

				case []int:
					for _, v9 := range nv {
						list1T = append(list1T, v9)
					}

				case []rune:
					for _, v9 := range nv {
						list1T = append(list1T, v9)
					}

				case []string:
					for _, v9 := range nv {
						list1T = append(list1T, v9)
					}

				case []interface{}:
					for _, v9 := range nv {
						list1T = append(list1T, v9)
					}

				}
			} else {
				// tk.Pl("not slice: %v", v)
				// tk.Pl("not slice value: %v", p.GetVarValue(v))
				list1T = append(list1T, p.GetVarValue(r, v))
			}

		}

		fmt.Printf(formatT+"\n", list1T...)

		return ""
	case 10422: // plNow
		list1T := []interface{}{}

		formatT := ""

		for i, v := range instrT.Params {
			if i == 0 {
				formatT = tk.ToStr(v.Value)
				continue
			}

			list1T = append(list1T, p.GetVarValue(r, v))
		}

		tk.PlNow(formatT, list1T...)

		return ""

	case 10430: // plv
		if instrT.ParamLen < 1 {
			tk.Plv(p.GetCurrentFuncContext(r).Tmp)
			return ""
		}

		s1 := p.GetVarValue(r, instrT.Params[0])

		tk.Plv(s1)

		return ""

	case 10433: // plvsr

		vs := p.ParamsToList(r, instrT, 0)

		tk.Plvsr(vs)

		return ""

	case 10440: // plErr
		if instrT.ParamLen < 1 {
			tk.PlErr(p.GetCurrentFuncContext(r).Tmp.(error))
			return ""
			// return p.Errf(r, "not enough parameters(参数不够)")
		}

		s1 := p.GetVarValue(r, instrT.Params[0]).(error)

		tk.PlErr(s1)

		return ""

	case 10441: // plErrX
		if instrT.ParamLen < 1 {
			tk.PlErrX(p.GetCurrentFuncContext(r).Tmp.(error))
			return ""
			// return p.Errf(r, "not enough parameters(参数不够)")
		}

		s1 := p.GetVarValue(r, instrT.Params[0]).(error)

		tk.PlErrX(s1)

		return ""

	case 10450: // plErrStr
		if instrT.ParamLen < 1 {
			tk.PlErrString(tk.ToStr(p.GetCurrentFuncContext(r).Tmp))
			return ""
			// return p.Errf(r, "not enough parameters(参数不够)")
		}

		s1 := p.GetVarValue(r, instrT.Params[0]).(string)

		tk.PlErrString(s1)

		return ""

	case 10460: // spr
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))

		v3 := p.ParamsToList(r, instrT, 2)

		errT := p.SetVar(r, pr, fmt.Sprintf(v2, v3...))

		if errT != nil {
			return p.Errf(r, "变量赋值错误：%v", errT)
		}

		return ""
	case 10511: // scanf
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[0]))

		vs := p.ParamsToList(r, instrT, 1)

		_, errT := fmt.Scanf(v1, vs...)

		if errT != nil {
			return p.Errf(r, "扫描数据失败：%v", errT)
		}

		return ""
	case 10512: // sscanf
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[0]))

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))

		v3 := p.ParamsToList(r, instrT, 2)

		_, errT := fmt.Sscanf(v1, v2, v3...)

		if errT != nil {
			return p.Errf(r, "扫描数据失败：%v", errT)
		}

		return ""

	case 10810: // convert/转换
		// tk.Plv(instrT)
		if instrT.ParamLen < 2 {
			return p.Errf(r, "参数个数不够")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		if tk.IsError(v1) {
			return p.Errf(r, "参数错误")
		}

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		if tk.IsError(v2) {
			return p.Errf(r, "参数错误")
		}

		var v3 interface{}

		s2 := v2.(string)

		if s2 == "bool" {
			v3 = tk.ToBool(v1)
		} else if s2 == "int" {
			v3 = tk.ToInt(v1)
		} else if s2 == "byte" {
			v3 = tk.ToByte(v1)
		} else if s2 == "rune" {
			v3 = tk.ToRune(v1)
		} else if s2 == "float" {
			v3 = tk.ToFloat(v1)
		} else if s2 == "str" {
			// nv, ok := v1.(XieObject)

			// if ok {
			// 	v3 = nv.Call("toStr")
			// } else {
			v3 = tk.ToStr(v1)
			// }
		} else if s2 == "list" {

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
				return p.Errf(r, "无法处理的类型")
			}
		} else if s2 == "strList" {

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
				return p.Errf(r, "无法处理的类型")
			}
		} else if s2 == "byteList" {

			switch nv := v1.(type) {
			case string:
				v3 = []byte(nv)
			case []rune:
				v3 = []byte(string(nv))
			default:
				return p.Errf(r, "无法处理的类型")
			}
		} else if s2 == "runeList" {

			switch nv := v1.(type) {
			case string:
				v3 = []rune(nv)
			case []byte:
				v3 = []byte(string(nv))
			default:
				return p.Errf(r, "无法处理的类型")
			}
		} else if s2 == "map" {

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
				return p.Errf(r, "无法处理的类型")
			}
		} else if s2 == "strMap" {

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
				return p.Errf(r, "无法处理的类型")
			}
		} else if s2 == "time" {

			switch nv := v1.(type) {
			case time.Time:
				v3 = nv
			case string:
				v3 = tk.ToTime(nv, p.ParamsToList(r, instrT, v1p+2)...)
			default:
				tmps := tk.ToStr(v1)
				v3 = tk.ToTime(tmps, p.ParamsToList(r, instrT, v1p+2)...)
			}
		} else if s2 == "timeStr" {

			switch nv := v1.(type) {
			case time.Time:
				v3 = tk.FormatTime(nv, p.ParamsToStrs(r, instrT, v1p+2)...)
			case string:
				rs := tk.ToTime(nv)
				if tk.IsError(rs) {
					return p.Errf(r, "时间转换失败：%v", rs)
				}

				v3 = tk.FormatTime(rs.(time.Time), p.ParamsToStrs(r, instrT, v1p+2)...)
			default:
				rs := tk.ToTime(tk.ToStr(v1))
				if tk.IsError(rs) {
					return p.Errf(r, "时间转换失败：%v", rs)
				}

				v3 = tk.FormatTime(rs.(time.Time), p.ParamsToStrs(r, instrT, v1p+2)...)
			}
		} else if s2 == "tick" || s2 == "timeStamp" {

			switch nv := v1.(type) {
			case time.Time:
				v3 = tk.GetTimeStampMid(nv)
			default:
				p.Errf(r, "类型不匹配：%v", v1)
			}
		} else if s2 == "postData" || s2 == "url.Values" {

			switch nv := v1.(type) {
			case map[string]string:
				v3 = tk.MapToPostData(nv)
			case map[string]interface{}:
				v3 = tk.MapToPostDataI(nv)
			default:
				p.Errf(r, "类型不匹配：%v", v1)
			}
		} else {
			return p.Errf(r, "无法处理的类型")
		}

		p.SetVar(r, pr, v3)

		return ""

	case 10821: // hex
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		var v3 interface{}

		v3 = tk.DataToBytes(v1, "-endian=B")

		if tk.IsError(v3) {
			return p.Errf(r, "转换失败：%v", v3)
		}

		v3 = hex.EncodeToString(v3.([]byte))
		p.SetVar(r, pr, v3)

		return ""

	case 10822: // hexb
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		var v3 interface{}

		v3 = tk.DataToBytes(v1, "-endian=L")

		if tk.IsError(v3) {
			return p.Errf(r, "转换失败：%v", v3)
		}

		v3 = hex.EncodeToString(v3.([]byte))
		p.SetVar(r, pr, v3)

		return ""

	case 10823: // unhex
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, tk.HexToBytes(v1))

		return ""

	case 10824: // toHex
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, tk.ToHex(v1))

		return ""

	case 10825: // hexToByte
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		if len(v1) < 2 {
			v1 = "0" + v1
		}

		rs1 := tk.HexToBytes(v1)
		if rs1 == nil {
			return fmt.Errorf("failed to convert hex: %v", v1)
		}

		p.SetVar(r, pr, rs1[0])

		return ""

	case 10831: // toBool
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, tk.ToBool(v1))

		return ""

	case 10835: // toByte
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, tk.ToByte(v1))

		return ""

	case 10837: // toRune
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, tk.ToRune(v1))

		return ""

	case 10851: // toInt
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		if instrT.ParamLen > 2 {
			v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))

			p.SetVar(r, pr, tk.ToInt(v1, v2))

			return ""
		}

		p.SetVar(r, pr, tk.ToInt(v1))

		return ""

	case 10855: // toFloat
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		if instrT.ParamLen > 2 {
			v2 := tk.ToFloat(p.GetVarValue(r, instrT.Params[v1p+1]))

			p.SetVar(r, pr, tk.ToFloat(v1, v2))

			return ""
		}

		p.SetVar(r, pr, tk.ToFloat(v1))

		return ""

	case 10861: // toStr
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, tk.ToStr(v1))

		return ""

	case 10871: // toTime
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToTime(p.GetVarValue(r, instrT.Params[v1p]), p.ParamsToList(r, instrT, v1p+1)...)

		p.SetVar(r, pr, v1)

		return ""
	case 10891: // toAny
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		switch nv := v1.(type) {
		case reflect.Value:
			p.SetVar(r, pr, nv.Interface())
			return ""
		}

		p.SetVar(r, pr, interface{}(v1))

		return ""
	case 10910: // isErrStr
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p]).(string)

		var rsT bool

		errMsgT := ""

		if tk.IsErrStr(v1) {
			rsT = true
		} else {
			rsT = false
		}

		p.SetVar(r, pr, rsT)

		if instrT.ParamLen > 2 {
			if rsT {
				errMsgT = tk.GetErrStr(v1)
			}

			p.SetVar(r, instrT.Params[2], errMsgT)
		}

		return ""

	case 10915: // errStrf
		pr := instrT.Params[0]

		list1T := []interface{}{}

		formatT := ""

		for i, v := range instrT.Params {
			if i == 0 {
				continue
			}
			if i == 1 {
				formatT = tk.ToStr(v.Value)
				continue
			}

			list1T = append(list1T, p.GetVarValue(r, v))
		}

		p.SetVar(r, pr, tk.ErrStrf(formatT, list1T...))
		return ""

	case 10921: // getErrStr
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, tk.GetErrStr(v1))

		return ""

	case 10931: // checkErrStr
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[0]))

		if tk.IsErrStr(v1) {
			// tk.Pln(v1)
			return p.Errf(r, tk.GetErrStr(v1))
			// return "exit"
		}

		return ""
	case 10941: // isErr
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		var rsT bool

		errMsgT := ""

		if tk.IsError(v1) {
			rsT = true
		} else {
			rsT = false
		}

		p.SetVar(r, pr, rsT)

		if instrT.ParamLen > 2 {
			if rsT {
				errMsgT = v1.(error).Error()
			}

			p.SetVar(r, instrT.Params[2], errMsgT)
		}

		return ""
	case 10942: // getErrMsg
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v1v, ok := v1.(error)

		if !ok {
			p.SetVar(r, pr, fmt.Errorf("type not match: %T(%#v)", v1, v1))
			return ""
		}

		p.SetVar(r, pr, v1v.Error())

		return ""

	case 10943: // isErrX
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		if tk.IsErrX(v1) {
			if instrT.ParamLen > 2 {
				p2 := instrT.Params[v1p+1]
				p.SetVar(r, p2, tk.GetErrStrX(v1))
			}

			p.SetVar(r, pr, true)

			return ""
		}

		if instrT.ParamLen > 2 {
			p2 := instrT.Params[v1p+1]
			p.SetVar(r, p2, "")
		}

		p.SetVar(r, pr, false)

		return ""

	case 10945: // checkErrX
		// if instrT.ParamLen < 1 {
		// 	if tk.IsErrX(p.GetCurrentFuncContext(r).Tmp) {
		// 		// p.RunDeferUpToRoot()
		// 		return p.Errf(r, tk.GetErrStrX(p.GetCurrentFuncContext(r).Tmp))
		// 	}

		// 	return ""
		// }

		var v1 interface{}

		if instrT.ParamLen > 0 {
			v1 = p.GetVarValue(r, instrT.Params[0])
		} else {
			v1 = p.GetCurrentFuncContext(r).Tmp
		}

		// v1 := p.GetVarValue(r, instrT.Params[0])

		if tk.IsErrX(v1) {
			// p.RunDeferUpToRoot(r)
			return p.Errf(r, "%v", tk.GetErrStrX(v1))
		}

		return ""

	case 10947: // getErrStrX
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, tk.GetErrStrX(v1))

		return ""

	case 10949: // errf
		pr := instrT.Params[0]

		list1T := []interface{}{}

		formatT := ""

		for i, v := range instrT.Params {
			if i == 0 {
				continue
			}
			if i == 1 {
				formatT = tk.ToStr(v.Value)
				continue
			}

			list1T = append(list1T, p.GetVarValue(r, v))
		}

		p.SetVar(r, pr, tk.Errf(formatT, list1T...))
		return ""

	case 12001: // clear
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// var pr interface{} = -5
		v1p := 0

		// if instrT.ParamLen > 1 {
		// 	pr = instrT.Params[0]
		// 	v1p = 1
		// }

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		switch nv := v1.(type) {
		case *tk.ByteQueue:
			nv.Clear()

			return ""
		default:
			tk.Pl("clear %T(%v)", v1, v1)
			tk.ReflectCallMethod(v1, "Clear")

			return ""
		}

		return p.Errf(r, "不支持的类型(type not supported)：%T(%v)", v1, v1)

	case 12003: // close
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		switch nv := v1.(type) {
		case *os.File:
			errT := nv.Close()

			p.SetVar(r, pr, errT)

			return ""
		default:
			tk.Pl("close %T(%v)", v1, v1)
			rvr := tk.ReflectCallMethod(v1, "Close", p.ParamsToList(r, instrT, v1p+1)...)

			p.SetVar(r, pr, rvr)

			return ""
		}

		p.SetVar(r, pr, nil)

		return p.Errf(r, "不支持的类型(type not supported)：%T(%v)", v1, v1)

	case 20110: // writeResp
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0]).(http.ResponseWriter)

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))

		tk.WriteResponse(v1, v2)

		return ""

	case 20111: // setRespHeader
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0]).(http.ResponseWriter)

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))

		v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[2]))

		v1.Header().Set(v2, v3)

		return ""

	case 20112: // writeRespHeader
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0]).(http.ResponseWriter)

		v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[1]))

		v1.WriteHeader(v2)

		return ""

	case 20113: // getReqHeader
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p]).(*http.Request)

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		p.SetVar(r, pr, v1.Header.Get(v2))

		return ""

	case 20114: // genJsonResp/genResp
		if instrT.ParamLen < 4 {
			return p.Errf(r, "参数不够：%v", instrT.ParamLen)
		}

		pr := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1]).(*http.Request)

		v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[2]))

		v4 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))

		rsT := tk.GenerateJSONPResponseWithMore(v3, v4, v2, p.ParamsToStrs(r, instrT, 4)...)

		p.SetVar(r, pr, rsT)

		return ""

	case 20116: // serveFile
		if instrT.ParamLen < 3 {
			return p.Errf(r, "参数不够：%v", instrT.ParamLen)
		}

		// pr := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[0]).(http.ResponseWriter)

		v3 := p.GetVarValue(r, instrT.Params[1]).(*http.Request)

		v4 := tk.ToStr(p.GetVarValue(r, instrT.Params[2]))

		http.ServeFile(v2, v3, v4)

		return ""

	case 20121: // newMux
		var p1 interface{} = -5

		if instrT.ParamLen > 0 {
			p1 = instrT.Params[0]
		}

		p.SetVar(r, p1, http.NewServeMux())

		return ""

	case 20122: // setMuxHandler
		if instrT.ParamLen < 4 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0]).(*http.ServeMux)
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))
		v3 := p.GetVarValue(r, instrT.Params[2])
		v4 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))

		// var inputG interface{}

		// if instrT.ParamLen > 3 {
		// 	inputG = p.GetVarValue(r, instrT.Params[3])
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

			vmT := NewVMQuick()

			vmT.SetVar(nil, "paraMapG", paraMapT)
			vmT.SetVar(nil, "requestG", req)
			vmT.SetVar(nil, "responseG", res)
			vmT.SetVar(nil, "reqNameG", req.RequestURI)
			vmT.SetVar(nil, "inputG", v3)

			lrs := vmT.Load(nil, v4)

			if tk.IsError(lrs) {
				res.Write([]byte(fmt.Sprintf("操作失败：%v", tk.GetErrStrX(lrs))))
				return
			}

			rs := vmT.Run()

			if tk.IsError(rs) {
				res.Write([]byte(fmt.Sprintf("操作失败：%v", tk.GetErrStrX(rs))))
				return
			}

			toWriteT = tk.ToStr(rs)

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
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := p.GetVarValue(r, instrT.Params[0]).(*http.ServeMux)
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))
		v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[2]))
		// v4 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))

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
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))
		if !strings.HasPrefix(v2, ":") {
			v2 = ":" + v2
		}
		v3 := p.GetVarValue(r, instrT.Params[2]).(*http.ServeMux)

		ifGoT := tk.IfSwitchExists(p.ParamsToStrs(r, instrT, 3), "-go")

		if ifGoT {
			go http.ListenAndServe(v2, v3)
			p.SetVar(r, p1, "")

			return ""
		}

		errT := http.ListenAndServe(v2, v3)

		if errT != nil {
			p.SetVar(r, p1, fmt.Errorf("启动服务失败：%v", errT))
		} else {
			p.SetVar(r, p1, "")
		}

		return ""

	case 20153: // startHttpsServer
		if instrT.ParamLen < 5 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))
		if !strings.HasPrefix(v2, ":") {
			v2 = ":" + v2
		}
		v3 := p.GetVarValue(r, instrT.Params[2]).(*http.ServeMux)
		v4 := tk.ToStr(p.GetVarValue(r, instrT.Params[3]))
		v5 := tk.ToStr(p.GetVarValue(r, instrT.Params[4]))

		ifGoT := tk.IfSwitchExists(p.ParamsToStrs(r, instrT, 5), "-go")

		if ifGoT {
			go http.ListenAndServeTLS(v2, v4, v5, v3)
			p.SetVar(r, p1, "")

			return ""
		}

		errT := http.ListenAndServeTLS(v2, v4, v5, v3)

		if errT != nil {
			p.SetVar(r, p1, fmt.Errorf("启动服务失败：%v", errT))
		} else {
			p.SetVar(r, p1, "")
		}

		return ""

	case 20210: // getWeb
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1])

		listT := p.ParamsToList(r, instrT, 2)

		rs := tk.GetWeb(tk.ToStr(v2), listT...)

		p.SetVar(r, pr, rs)

		return ""

	case 20213: // getWebBytes
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1])

		listT := p.ParamsToList(r, instrT, 2)

		listT = append(listT, "-bytes")

		rs := tk.GetWeb(tk.ToStr(v2), listT...)

		p.SetVar(r, pr, rs)

		return ""

	case 20220: // downloadFile
		if instrT.ParamLen < 4 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v1p := 1

		v2 := p.GetVarValue(r, instrT.Params[v1p])
		v3 := p.GetVarValue(r, instrT.Params[v1p+1])
		v4 := p.GetVarValue(r, instrT.Params[v1p+2])

		vs := p.ParamsToStrs(r, instrT, v1p+3)

		if tk.IfSwitchExistsWhole(vs, "-progress") {
			fmt.Println()
			rs := tk.DownloadFileWithProgress(tk.ToStr(v2), tk.ToStr(v3), tk.ToStr(v4), func(i interface{}) interface{} {
				fmt.Printf("\rprogress: %v                  ", tk.IntToKMGT(i))
				return ""
			}, vs...)

			p.SetVar(r, pr, rs)

			return ""
		}

		rs := tk.DownloadFile(tk.ToStr(v2), tk.ToStr(v3), tk.ToStr(v4), vs...)

		p.SetVar(r, pr, rs)

		return ""

	case 20291: // getResource
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		// ResourceLockG.Lock()

		textT := tk.SafelyGetStringForKeyWithDefault(ResourceG, v2, "")

		p.SetVar(r, pr, strings.ReplaceAll(textT, "~~~", "`"))

		return ""

	case 20292: // getResourceRaw
		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		// ResourceLockG.Lock()

		textT := tk.SafelyGetStringForKeyWithDefault(ResourceG, v2, "")

		// ResourceLockG.Unlock()

		p.SetVar(r, pr, textT)

		return ""

	case 20293: // getResourceList
		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		// ResourceLockG.Lock()

		aryT := make([]string, 0, len(ResourceG))
		for k, _ := range ResourceG {
			aryT = append(aryT, k)
		}
		// ResourceLockG.Unlock()

		p.SetVar(r, pr, aryT)

		return ""

	case 20310: // htmlToText
		// var v2 []string
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		// if instrT.ParamLen < 3 {
		// 	v2 = []string{}
		// } else {
		// 	v2 = p.GetVarValue(r, instrT.Params[2]).([]string)
		// }

		v1 := p.GetVarValue(r, instrT.Params[1])

		v2 := p.ParamsToStrs(r, instrT, 2)

		rs := tk.HTMLToText(tk.ToStr(v1), v2...)

		p.SetVar(r, pr, rs)

		return ""

	case 20411: // regReplace/regReplaceAllStr
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
		v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))

		rs := tk.RegReplaceX(v1, v2, v3)

		p.SetVar(r, pr, rs)

		return ""
	case 20421: // regFindAll
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := p.GetVarValue(r, instrT.Params[v1p+2])

		rs := tk.RegFindAllX(tk.ToStr(v1), tk.ToStr(v2), tk.ToInt(v3, 0))

		p.SetVar(r, pr, rs)

		return ""

	case 20423: // regFind
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v3 := p.GetVarValue(r, instrT.Params[v1p+2])

		rs := tk.RegFindFirstX(tk.ToStr(v1), tk.ToStr(v2), tk.ToInt(v3, 0))

		p.SetVar(r, pr, rs)

		return ""

	case 20425: // regFindIndex
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		rs1, rs2 := tk.RegFindFirstIndexX(tk.ToStr(v1), tk.ToStr(v2))

		p.SetVar(r, pr, []int{rs1, rs2})

		return ""

	case 20431: // regMatch
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		rs := tk.RegMatchX(tk.ToStr(v1), tk.ToStr(v2))

		p.SetVar(r, pr, rs)

		return ""

	case 20441: // regContains
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		rs := tk.RegContainsX(tk.ToStr(v1), tk.ToStr(v2))

		p.SetVar(r, pr, rs)

		return ""

	case 20443: // regContainsIn
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2s := p.ParamsToStrs(r, instrT, v1p+1)

		rs := tk.RegContainsIn(tk.ToStr(v1), v2s...)

		p.SetVar(r, pr, rs)

		return ""

	case 20445: // regCount
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		rs := tk.RegFindAllIndexX(tk.ToStr(v1), tk.ToStr(v2))

		p.SetVar(r, pr, len(rs))

		return ""

	case 20451: // regSplit
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		s1 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))

		s2 := tk.ToStr(p.GetVarValue(r, instrT.Params[2]))

		countT := -1

		if instrT.ParamLen > 3 {
			countT = tk.ToInt(p.GetVarValue(r, instrT.Params[3]))
		}

		listT := tk.RegSplitX(s1, s2, countT)

		p.SetVar(r, pr, listT)

		return ""

	case 20491: // regQuote
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, regexpx.QuoteMeta(v1))
		return ""

	case 20501: // sleep
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v1 := tk.ToFloat(p.GetVarValue(r, instrT.Params[0]))

		tk.Sleep(v1)

		return ""

	case 20511: // getClipText
		var pr interface{} = -5

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		strT := tk.GetClipText()

		if tk.IsErrX(strT) {
			p.SetVar(r, pr, fmt.Errorf("%v", tk.GetErrStrX(strT)))
			return
		}

		p.SetVar(r, pr, strT)
		return ""

	case 20512: // setClipText
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rsT := tk.SetClipText(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])))

		p.SetVar(r, pr, rsT)

		return ""
	case 20521: // getEnv
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rsT := tk.GetEnv(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])))

		p.SetVar(r, pr, rsT)

		return ""

	case 20522: // setEnv
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rsT := tk.SetEnv(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])), tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1])))

		p.SetVar(r, pr, rsT)

		return ""

	case 20523: // removeEnv
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rsT := os.Unsetenv(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])))

		p.SetVar(r, pr, rsT)

		return ""

	case 20601: // systemCmd
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		optsA := p.ParamsToStrs(r, instrT, v1p+1)

		// tk.Pln(v1, ",", optsA)

		p.SetVar(r, pr, tk.SystemCmd(v1, optsA...))

		return ""
	case 20603: // openWithDefault
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := tk.RunWinFileWithSystemDefault(v1)

		if rsT != "" {
			rsT = tk.GenerateErrorString(rsT)
		}

		p.SetVar(r, pr, rsT)

		return ""

	case 20901: // getOSName
		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		p.SetVar(r, pr, runtime.GOOS)

		return ""

	case 21101: // loadText
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		fcT, errT := tk.LoadStringFromFileE(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])))

		if errT != nil {
			p.SetVar(r, pr, errT)
		} else {
			p.SetVar(r, pr, fcT)
		}

		return ""

	case 21103: // saveText
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rsT := tk.SaveStringToFile(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])), tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1])))

		if rsT != "" {
			p.SetVar(r, pr, fmt.Errorf(tk.GetErrStr(rsT)))
		} else {
			p.SetVar(r, pr, "")
		}

		return ""
	case 21105: // loadBytes
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		fcT := tk.LoadBytesFromFile(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])))

		p.SetVar(r, pr, fcT)

		return ""

	case 21106: // saveBytes
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		v1b, ok := v1.([]byte)

		var rsT interface{}

		if ok {
			rsT = tk.SaveBytesToFileE(v1b, v2)

			p.SetVar(r, pr, rsT)
			return ""
		}

		v1buf, ok := v1.(*bytes.Buffer)

		if ok {
			rsT = tk.SaveBytesToFileE(v1buf.Bytes(), v2)

			p.SetVar(r, pr, rsT)
			return ""
		}

		// p.SetVar(r, pr, fmt.Errorf())

		return p.Errf(r, "无法处理的类型：%T", v1)

	case 21107: // loadBytesLimit
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		fcT := tk.LoadBytesFromFile(tk.ToStr(p.GetVarValue(r, instrT.Params[1])), tk.ToInt(p.GetVarValue(r, instrT.Params[2])))

		p.SetVar(r, pr, fcT)

		return ""

	case 21111: // appendText
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rsT := tk.AppendStringToFile(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])), tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1])))

		if rsT != "" {
			p.SetVar(r, pr, fmt.Errorf(tk.GetErrStr(rsT)))
		} else {
			p.SetVar(r, pr, "")
		}

		return ""

	case 21201: // writeStr
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		switch nv := v1.(type) {
		case string:
			p.SetVar(r, pr, nv+v2)
			return ""
		case io.StringWriter:
			n, err := nv.WriteString(v2)

			if err != nil {
				p.SetVar(r, pr, err)
				return ""
			}

			p.SetVar(r, pr, n)
			return ""
		default:
			p.SetVar(r, pr, fmt.Errorf("无法处理的类型：%T(%v)", v1, v1))
			return ""

		}

		return ""

	case 21501: // createFile
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		optsA := p.ParamsToStrs(r, instrT, v1p+1)

		if tk.IfSwitchExistsWhole(optsA, "-return") {
			if !tk.IfSwitchExistsWhole(optsA, "-overwrite") && tk.IfFileExists(v1) {
				p.SetVar(r, pr, fmt.Errorf("文件已存在"))
				return ""
			}

			fileT, errT := os.Create(v1)

			if errT != nil {
				p.SetVar(r, pr, errT)
				return ""
			}

			p.SetVar(r, pr, fileT)
			return ""
		}

		errT := tk.CreateFile(v1, optsA...)

		p.SetVar(r, pr, errT)

		return ""

	case 21503: // openFile
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		vs := p.ParamsToStrs(r, instrT, v1p+1)

		if tk.IfSwitchExistsWhole(vs, "-read") {
			fileT, errT := os.Open(v1)

			if errT != nil {
				p.SetVar(r, pr, errT)

				return ""
			}

			p.SetVar(r, pr, fileT)

			return ""

		}

		fileT := tk.OpenFile(v1, vs...)

		p.SetVar(r, pr, fileT)

		return ""

	case 21505: // openFileForRead
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		fileT, errT := os.Open(v1)

		if errT != nil {
			p.SetVar(r, pr, fileT)
			return ""
		}

		p.SetVar(r, pr, fileT)

		return ""

	case 21507: // closeFile
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p]).(*os.File)

		errT := v1.Close()

		p.SetVar(r, pr, errT)

		return ""

	case 21521: // readByte
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		switch nv := v1.(type) {
		case *bufio.Reader:
			byteT, errT := nv.ReadByte()

			if errT != nil {
				p.SetVar(r, pr, errT)
			} else {
				p.SetVar(r, pr, byteT)
			}

			return ""
		case io.ByteReader:
			byteT, errT := nv.ReadByte()

			if errT != nil {
				p.SetVar(r, pr, errT)
			} else {
				p.SetVar(r, pr, byteT)
			}

			return ""
		case io.Reader:
			bufT := make([]byte, 1)
			cntT, errT := nv.Read(bufT)

			if errT != nil {
				if cntT > 0 {
					p.SetVar(r, pr, bufT[0])
				} else {
					p.SetVar(r, pr, errT)
				}

			} else {
				p.SetVar(r, pr, bufT[0])
			}

			return ""
		default:
			tk.Pl("readByte %T(%v)", v1, v1)
			rvr := tk.ReflectCallMethod(v1, "ReadByte", p.ParamsToList(r, instrT, v1p+1)...)

			p.SetVar(r, pr, rvr)

			return ""
		}

		p.SetVar(r, pr, nil)

		return p.Errf(r, "不支持的类型(type not supported)：%T(%v)", v1, v1)

	case 21525: // readBytesN
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))

		switch nv := v1.(type) {
		case io.Reader:
			bufT := make([]byte, v2)

			cntT, errT := nv.Read(bufT)

			if errT != nil {
				p.SetVar(r, pr, errT)
				return ""
			}

			if cntT == v2 {
				p.SetVar(r, pr, bufT)
				return ""
			}

			leftT := v2 - cntT

			for {
				buf1T := make([]byte, leftT)

				cntT, errT = nv.Read(buf1T)

				if errT != nil {
					p.SetVar(r, pr, errT)
					return ""
				}

				bufT = append(bufT[:v2-leftT], buf1T[:cntT]...)

				leftT = leftT - cntT

				if leftT <= 0 {
					p.SetVar(r, pr, bufT)
					return ""
				}
			}

			p.SetVar(r, pr, fmt.Errorf("failed to read %v bytes: %#v", v2, bufT))

			return ""
		default:
			p.SetVar(r, pr, fmt.Errorf("type not supported(不支持的类型): %T(%v)", v1, v1))

			return ""
		}

		p.SetVar(r, pr, fmt.Errorf("type not supported(不支持的类型): %T(%v)", v1, v1))

		return ""

	case 21531: // writeByte
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// var pr interface{} = -5
		// v1p := 0

		// if instrT.ParamLen > 1 {
		pr := instrT.Params[0]
		v1p := 1
		// }

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v2 := tk.ToByte(p.GetVarValue(r, instrT.Params[v1p+1]))

		nv1, ok := v1.(*tk.ByteQueue)

		if ok {
			nv1.Push(v2)
			p.SetVar(r, pr, nil)
			return ""
		}

		w1, ok := v1.(io.Writer)

		if !ok {
			return p.Errf(r, "type not supported(不支持的类型): %T(%v)", v1, v1)
		}

		w2, ok := v1.(*bufio.Writer)

		if !ok {
			w2 = nil
		}

		_, err1 := w1.Write([]byte{v2})

		if err1 != nil {
			p.SetVar(r, pr, err1)
			return nil
		}

		if w2 != nil {
			w2.Flush()
		}

		p.SetVar(r, pr, nil)

		return ""

	case 21533: // writeBytes
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// var pr interface{} = -5
		// v1p := 0

		// if instrT.ParamLen > 1 {
		pr := instrT.Params[0]
		v1p := 1
		// }

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		w1, ok := v1.(io.Writer)

		if !ok {
			return p.Errf(r, "type not supported(不支持的类型): %T(%v)", v1, v1)
		}

		w2, ok := v1.(*bufio.Writer)

		if !ok {
			w2 = nil
		}

		vs := p.ParamsToList(r, instrT, v1p+1)

		var cntAllT int = 0

		for _, jjv := range vs {
			switch nv := jjv.(type) {
			case byte:
				cnt1, err1 := w1.Write([]byte{nv})

				if err1 != nil {
					p.SetVar(r, pr, err1)
					return nil
				}

				cntAllT += cnt1
			case string:
				cnt1, err1 := w1.Write(tk.HexToBytes(nv))

				if err1 != nil {
					p.SetVar(r, pr, err1)
					return nil
				}

				cntAllT += cnt1
			case []byte:
				cnt1, err1 := w1.Write(nv)

				if err1 != nil {
					p.SetVar(r, pr, err1)
					return nil
				}

				cntAllT += cnt1
			default:
				p.SetVar(r, pr, p.Errf(r, "type not supported(不支持的类型): %T(%v)", v1, v1))
				// tk.Pl("write %T(%v)", v1, v1)
				// rvr := tk.ReflectCallMethod(v1, "Write", p.ParamsToList(r, instrT, v1p+1)...)

				// p.SetVar(r, pr, rvr)

				return ""
			}
		}

		if w2 != nil {
			w2.Flush()
		}

		p.SetVar(r, pr, cntAllT)

		return ""

	case 21541: // flush
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		switch nv := v1.(type) {
		case *bufio.Writer:
			p.SetVar(r, pr, nv.Flush())

			return ""
		default:
			tk.Pl("flush %T(%v)", v1, v1)
			rvr := tk.ReflectCallMethod(v1, "Flush", p.ParamsToList(r, instrT, v1p+1)...)

			p.SetVar(r, pr, rvr)

			return ""
		}

		p.SetVar(r, pr, nil)

		return p.Errf(r, "不支持的类型(type not supported)：%T(%v)", v1, v1)

	case 21601: // cmpBinFile
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		f1 := tk.ToStr(p.GetVarValue(r, instrT.Params[1]))
		f2 := tk.ToStr(p.GetVarValue(r, instrT.Params[2]))

		optsT := p.ParamsToStrs(r, instrT, 3)

		if tk.IfSwitchExistsWhole(optsT, "-identical") {
			buf1 := tk.LoadBytesFromFile(f1)

			if tk.IsError(buf1) {
				return p.Errf(r, "加载文件（%v）失败：%v", f1, buf1)
			}

			buf2 := tk.LoadBytesFromFile(f2)

			if tk.IsError(buf2) {
				return p.Errf(r, "加载文件（%v）失败：%v", f2, buf2)
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
					p.SetVar(r, pr, false)
					return ""
				} else {
					c1 = int(realBuf1[i])
				}

				if i >= len2 {
					p.SetVar(r, pr, false)
					return ""
				} else {
					c2 = int(realBuf2[i])
				}

				if c1 != c2 {
					p.SetVar(r, pr, false)
					return ""
				}

			}

		} else {
			buf1 := tk.LoadBytesFromFile(f1)

			if tk.IsError(buf1) {
				return p.Errf(r, "加载文件（%v）失败：%v", f1, buf1)
			}

			buf2 := tk.LoadBytesFromFile(f2)

			if tk.IsError(buf2) {
				return p.Errf(r, "加载文件（%v）失败：%v", f2, buf2)
			}

			realBuf1 := buf1.([]byte)
			realBuf2 := buf2.([]byte)

			p.SetVar(r, pr, tk.CompareBytes(realBuf1, realBuf2))

			return ""
		}

		p.SetVar(r, pr, true)

		return ""

	case 21701: // fileExists
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := tk.IfFileExists(v1)

		p.SetVar(r, pr, rsT)

		return ""

	case 21702: // isDir
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := tk.IsDirectory(v1)

		p.SetVar(r, pr, rsT)

		return ""

	case 21703: // getFileSize
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT, errT := tk.GetFileSize(v1)

		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		p.SetVar(r, pr, int(rsT))

		return ""

	case 21705: // getFileInfo
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT, errT := tk.GetFileInfo(v1)

		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		p.SetVar(r, pr, rsT)

		return ""

	case 21801: // removeFile
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		if instrT.ParamLen > 2 {
			optsA := p.ParamsToStrs(r, instrT, 2)

			if tk.IfSwitchExistsWhole(optsA, "-dry") {
				tk.Pl("模拟删除 %v", v1)

				p.SetVar(r, pr, nil)

				return ""
			}
		}

		rsT := tk.RemoveFile(v1)

		p.SetVar(r, pr, rsT)

		return ""

	case 21803: // renameFile
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		if instrT.ParamLen > 3 {
			optsA := p.ParamsToStrs(r, instrT, 3)

			p.SetVar(r, pr, tk.RenameFile(v1, v2, optsA...))

			return ""

		}

		p.SetVar(r, pr, tk.RenameFile(v1, v2))

		return ""

	case 21805: // copyFile
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		optsA := p.ParamsToStrs(r, instrT, 3)

		p.SetVar(r, pr, tk.RenameFile(v1, v2, optsA...))

		return ""

	case 21901: // genFileList/getFileList
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr = instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		paramsT := p.ParamsToStrs(r, instrT, 2)

		rsT := tk.GetFileList(v1, paramsT...)

		p.SetVar(r, pr, rsT)

		return ""

	case 21902: // joinPath
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		rsT := filepath.Join(p.ParamsToStrs(r, instrT, 1)...)

		p.SetVar(r, pr, rsT)

		return ""

	case 21905: // getCurDir

		var pr any = -5

		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		rsT := tk.GetCurrentDir()

		p.SetVar(r, pr, rsT)

		return ""

	case 21906: // setCurDir
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr any = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		dirT := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, tk.SetCurrentDir(dirT))

		return ""

	case 21907: // getAppDir

		var pr any = -5

		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		rsT := tk.GetApplicationPath()

		p.SetVar(r, pr, rsT)

		return ""

	case 21908: // getConfigDir

		// pr := -5
		// v1p := 0

		// if instrT.ParamLen > 1 {
		pr := instrT.Params[0]
		v1p := 1
		// }

		vs := p.ParamsToStrs(r, instrT, v1p)

		baseNameT := p.GetSwitchVarValue(r, vs, "-base=", "qx")

		rsT, errT := tk.EnsureBasePath(baseNameT)

		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		p.SetVar(r, pr, rsT)
		return ""
	case 21910: // extractFileName
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := filepath.Base(v1)

		p.SetVar(r, pr, rsT)

		return ""

	case 21911: // extractFileExt
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := filepath.Ext(v1)

		p.SetVar(r, pr, rsT)

		return ""

	case 21912: // extractFileDir
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := filepath.Dir(v1)

		p.SetVar(r, pr, rsT)

		return ""

	case 21915: // extractPathRel
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		rsT, errT := filepath.Rel(v2, v1)

		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		p.SetVar(r, pr, rsT)
		return ""

	case 21921: // ensureMakeDirs
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := tk.EnsureMakeDirs(v1)

		p.SetVar(r, pr, rsT)

		return ""
	case 22001: // getInput
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := tk.GetInputf(v1, p.ParamsToList(r, instrT, v1p+1)...)

		p.SetVar(r, pr, rsT)

		return ""

	case 22003: // getPassword
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := tk.GetInputPasswordf(v1, p.ParamsToList(r, instrT, v1p+1)...)

		p.SetVar(r, pr, rsT)

		return ""

	case 22101: // toJson
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = instrT.Params[0]

		argsT := p.ParamsToStrs(r, instrT, 2)

		vT := tk.ToJSONX(p.GetVarValue(r, instrT.Params[1]), argsT...)

		p.SetVar(r, pr, vT)

		return ""

	case 22102: // fromJson
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		vT, errT := tk.FromJSON(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])))

		if errT != nil {
			vT = errT
		}

		p.SetVar(r, pr, vT)

		return ""

	case 22201: // toXml
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var p1 interface{} = instrT.Params[0]

		argsT := p.ParamsToList(r, instrT, 2)

		vT := tk.ToXML(p.GetVarValue(r, instrT.Params[1]), argsT...)

		p.SetVar(r, p1, vT)

		return ""

	case 23000: // randomize
		if instrT.ParamLen > 0 {
			tk.Randomize(tk.ToInt(p.GetVarValue(r, instrT.Params[0])))
		} else {
			tk.Randomize()
		}

		return ""

	case 23001: // getRandomInt
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		minT := 0
		maxT := tk.MAX_INT

		pr := instrT.Params[0]

		if instrT.ParamLen > 2 {
			minT = tk.ToInt(p.GetVarValue(r, instrT.Params[1]))
			maxT = tk.ToInt(p.GetVarValue(r, instrT.Params[2]))
		} else {
			maxT = tk.ToInt(p.GetVarValue(r, instrT.Params[1]))
		}

		rs := tk.GetRandomIntInRange(minT, maxT)

		p.SetVar(r, pr, rs)

		return ""

	case 23003: // genRandomFloat
		var pr interface{} = -5
		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
		}

		p.SetVar(r, pr, rand.Float64())

		return ""

	case 23101: // genRandomStr
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]

		listT := p.ParamsToStrs(r, instrT, 1)

		rs := tk.GenerateRandomStringX(listT...)

		p.SetVar(r, p1, rs)

		return ""

	case 24101: // md5
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, tk.MD5Encrypt(tk.ToStr(v1)))

		return ""

	case 24201: // simpleEncode
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		var v2 byte = '_'
		if instrT.ParamLen > 2 {
			v2 = p.GetVarValue(r, instrT.Params[2]).(byte)
		}

		rsT := tk.EncodeStringCustomEx(v1, v2)

		p.SetVar(r, pr, rsT)

		return ""

	case 24203: // simpleDecode
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		var v2 byte = '_'
		if instrT.ParamLen > 2 {
			v2 = p.GetVarValue(r, instrT.Params[2]).(byte)
		}

		rsT := tk.DecodeStringCustom(v1, v2)

		p.SetVar(r, pr, rsT)

		return ""

	case 24301: // urlEncode
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		var v2 string = ""
		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(r, instrT.Params[2]))
		}

		if v2 == "-method=x" {
			p.SetVar(r, pr, tk.UrlEncode2(v1))
		} else {
			p.SetVar(r, pr, tk.UrlEncode(v1))
		}

		return ""

	case 24303: // urlDecode
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rStrT, errT := url.QueryUnescape(v1)
		if errT != nil {
			p.SetVar(r, pr, errT)
		} else {
			p.SetVar(r, pr, rStrT)
		}

		return ""

	case 24401: // base64Encode
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		p.SetVar(r, pr, tk.ToBase64(v1))

		return ""

	case 24403: // base64Decode
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rStrT, errT := base64.StdEncoding.DecodeString(v1)
		if errT != nil {
			p.SetVar(r, pr, errT)
		} else {
			p.SetVar(r, pr, rStrT)
		}

		return ""

	case 24501: // htmlEncode
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, tk.EncodeHTML(v1))

		return ""

	case 24503: // htmlDecode
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, tk.DecodeHTML(v1))

		return ""

	case 24601: // hexEncode
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, tk.StrToHex(v1))

		return ""

	case 24603: // hexDecode
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		if strings.HasPrefix(v1, "HEX_") {
			v1 = v1[4:]
		}

		p.SetVar(r, pr, tk.HexToStr(v1))

		return ""

	case 24801: // toUtf8/toUTF8
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		var v2 string = ""

		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
		}

		var rs interface{}

		switch nv := v1.(type) {
		case string:
			rs = tk.ConvertStringToUTF8(nv, v2)
		case []byte:
			rs = tk.ConvertToUTF8(nv, v2)
		default:
			return p.Errf(r, "参数类型错误：%T", v1)
		}

		p.SetVar(r, pr, rs)

		return ""

	case 25101: // encryptText
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		v2 := ""
		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
		}

		rsT := tk.EncryptStringByTXDEF(v1, v2)

		if tk.IsErrStr(rsT) {
			p.SetVar(r, pr, fmt.Errorf(tk.GetErrStr(rsT)))
		} else {
			p.SetVar(r, pr, rsT)
		}

		return ""

	case 25103: // decryptText
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		v2 := ""
		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
		}

		rsT := tk.DecryptStringByTXDEF(v1, v2)

		if tk.IsErrStr(rsT) {
			p.SetVar(r, pr, fmt.Errorf(tk.GetErrStr(rsT)))
		} else {
			p.SetVar(r, pr, rsT)
		}

		return ""

	case 25201: // encryptData
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p]).([]byte)

		v2 := ""
		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(r, instrT.Params[2]))
		}

		rsT := tk.EncryptDataByTXDEF(v1, v2)

		p.SetVar(r, pr, rsT)

		return ""

	case 25203: // decryptData
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p]).([]byte)

		v2 := ""
		if instrT.ParamLen > 2 {
			v2 = tk.ToStr(p.GetVarValue(r, instrT.Params[2]))
		}

		rsT := tk.DecryptDataByTXDEF(v1, v2)

		p.SetVar(r, pr, rsT)

		return ""
	case 26001: // compress
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		// bufT, ok := v1.([]byte)

		// if !ok {
		// 	p.SetVar(r, pr, fmt.Errorf("failed to compress, unsupported type: %T(%v)", v1, v1))
		// }

		rsT := tk.Compress(v1)

		p.SetVar(r, pr, rsT)

		return ""
	case 26002: // uncompress
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		// bufT, ok := v1.([]byte)

		// if !ok {
		// 	p.SetVar(r, pr, fmt.Errorf("failed to uncompress, unsupported type: %T(%v)", v1, v1))
		// }

		rsT := tk.Uncompress(v1)

		p.SetVar(r, pr, rsT)

		return ""
	case 26011: // compressText
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := tk.CompressText(v1)

		p.SetVar(r, pr, rsT)

		return ""
	case 26012: // uncompressText
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rsT := tk.UncompressText(v1)

		p.SetVar(r, pr, rsT)

		return ""
	case 27001: // getRandomPort
		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		listener, errT := net.Listen("tcp", ":0")
		if errT != nil {
			return p.Errf(r, "获取随机端口失败：%v", errT)
		}

		portT := listener.Addr().(*net.TCPAddr).Port
		// fmt.Println("Using port:", portT)
		listener.Close()

		p.SetVar(r, pr, portT)

		return ""
	case 27101: // listen
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}
		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rsT := tk.Listen(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])), tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1])))

		p.SetVar(r, pr, rsT)
		return ""

	case 27105: // accept
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}
		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		v1c, ok := v1.(net.Listener)

		if !ok {
			p.SetVar(r, pr, fmt.Errorf("failed to accept, invalid type: %T(%v)", v1, v1))
			return ""
		}

		rsT, errT := v1c.Accept()

		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		p.SetVar(r, pr, rsT)
		return ""

	case 32101: // dbConnect
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		dbT := sqltk.ConnectDBX(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])), tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1])))

		p.SetVar(r, pr, dbT)

		return ""

	case 32102: // dbClose
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		errT := sqltk.CloseDBX(p.GetVarValue(r, instrT.Params[v1p]).(*sql.DB))

		p.SetVar(r, pr, errT)

		return ""

	case 32103: // dbQuery
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1])

		v3 := p.GetVarValue(r, instrT.Params[2])

		listT := p.ParamsToList(r, instrT, 3)

		rs := sqltk.QueryDBX(v2.(*sql.DB), tk.ToStr(v3), listT...)

		if tk.IsError(rs) {
			p.SetVar(r, p1, rs)
		} else {
			p.SetVar(r, p1, rs)
		}

		return ""

	case 32104: // dbQueryMap
		if instrT.ParamLen < 4 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1])

		v3 := p.GetVarValue(r, instrT.Params[2])

		v4 := p.GetVarValue(r, instrT.Params[3])

		listT := p.ParamsToList(r, instrT, 4)

		rs := sqltk.QueryDBMapX(v2.(*sql.DB), tk.ToStr(v3), tk.ToStr(v4), listT...)

		if tk.IsError(rs) {
			p.SetVar(r, p1, rs)
		} else {
			p.SetVar(r, p1, rs)
		}

		return ""

	case 32105: // dbQueryRecs
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1])

		v3 := p.GetVarValue(r, instrT.Params[2])

		listT := p.ParamsToList(r, instrT, 3)

		rs := sqltk.QueryDBRecsX(v2.(*sql.DB), tk.ToStr(v3), listT...)

		if tk.IsError(rs) {
			p.SetVar(r, p1, rs)
		} else {
			p.SetVar(r, p1, rs)
		}

		return ""

	case 32106: // dbQueryCount
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1])

		v3 := p.GetVarValue(r, instrT.Params[2])

		listT := p.ParamsToList(r, instrT, 3)

		rs := sqltk.QueryCountX(v2.(*sql.DB), tk.ToStr(v3), listT...)

		if tk.IsError(rs) {
			p.SetVar(r, p1, rs)
		} else {
			p.SetVar(r, p1, rs)
		}

		return ""

	case 32107: // dbQueryFloat
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1])

		v3 := p.GetVarValue(r, instrT.Params[2])

		listT := p.ParamsToList(r, instrT, 3)

		rs := sqltk.QueryFloatX(v2.(*sql.DB), tk.ToStr(v3), listT...)

		if tk.IsError(rs) {
			p.SetVar(r, p1, rs)
		} else {
			p.SetVar(r, p1, rs)
		}

		return ""

	case 32108: // dbQueryString
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1])

		v3 := p.GetVarValue(r, instrT.Params[2])

		listT := p.ParamsToList(r, instrT, 3)

		rs := sqltk.QueryStringX(v2.(*sql.DB), tk.ToStr(v3), listT...)

		if tk.IsError(rs) {
			p.SetVar(r, p1, rs)
		} else {
			p.SetVar(r, p1, rs)
		}

		return ""

	case 32109: // dbQueryMapArray
		if instrT.ParamLen < 4 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1])

		v3 := p.GetVarValue(r, instrT.Params[2])

		v4 := p.GetVarValue(r, instrT.Params[3])

		listT := p.ParamsToList(r, instrT, 4)

		rs := sqltk.QueryDBMapArrayX(v2.(*sql.DB), tk.ToStr(v3), tk.ToStr(v4), listT...)

		if tk.IsError(rs) {
			p.SetVar(r, p1, rs)
		} else {
			p.SetVar(r, p1, rs)
		}

		return ""

	case 32111: // dbExec
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		p1 := instrT.Params[0]

		v2 := p.GetVarValue(r, instrT.Params[1])

		v3 := p.GetVarValue(r, instrT.Params[2])

		listT := p.ParamsToList(r, instrT, 3)

		rs := sqltk.ExecDBX(v2.(*sql.DB), tk.ToStr(v3), listT...)

		if tk.IsError(rs) {
			p.SetVar(r, p1, rs)
		} else {
			nv := rs.([]int64)
			p.SetVar(r, p1, []interface{}{int(nv[0]), int(nv[1])})
		}

		return ""

	case 40001: // renderMarkdown
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, tk.RenderMarkdown(v1))

		return ""

	case 41101: // pngEncode
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := p.GetVarValue(r, instrT.Params[v1p+1]).(image.Image)

		v1w, ok := v1.(io.Writer)
		if ok {
			errT := png.Encode(v1w, v2)

			p.SetVar(r, pr, errT)
			return ""

		}

		v1s := tk.ToStr(v1)

		fileT, errT := os.Create(v1s)

		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		defer fileT.Close()

		errT = png.Encode(fileT, v2)

		p.SetVar(r, pr, errT)
		return ""

	case 41103: // jpegEnocde/jpgEncode
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := p.GetVarValue(r, instrT.Params[v1p]).(io.Writer)
		v2 := p.GetVarValue(r, instrT.Params[v1p+1]).(image.Image)

		vs := p.ParamsToStrs(r, instrT, v1p+2)

		qualityT := tk.ToInt(tk.GetSwitch(vs, "-quality=", "90"), 90)

		if qualityT < 1 || qualityT > 100 {
			return p.Errf(r, "质量参数错误（1-100）：%v", qualityT)
		}

		v1w, ok := v1.(io.Writer)
		if ok {
			errT := jpeg.Encode(v1w, v2, &jpeg.Options{Quality: qualityT})

			p.SetVar(r, pr, errT)
			return ""

		}

		v1s := tk.ToStr(v1)

		fileT, errT := os.Create(v1s)

		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		defer fileT.Close()

		errT = jpeg.Encode(fileT, v2, &jpeg.Options{Quality: qualityT})

		p.SetVar(r, pr, errT)
		return ""

	case 45001: // getActiveDisplayCount
		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		p.SetVar(r, pr, screenshot.NumActiveDisplays())

		return ""

	case 45011: // getScreenResolution
		pr := instrT.Params[0]

		vs := p.ParamsToStrs(r, instrT, 1)

		formatT := p.GetSwitchVarValue(r, vs, "-format=", "")

		idxStrT := p.GetSwitchVarValue(r, vs, "-index=", "0")

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
		p.SetVar(r, pr, vr)

		return ""

	case 45021: // captureDisplay
		// var pr interface{} = -5
		// v1p := 0

		// if instrT.ParamLen > 0 {
		pr := instrT.Params[0]
		v1p := 1
		// }

		v1 := tk.ToInt(p.GetVarValue(r, GetVarRefFromArray(instrT.Params, v1p)), 0)

		imageA, errT := screenshot.CaptureDisplay(v1)

		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		p.SetVar(r, pr, imageA)
		return ""

	case 45023: // captureScreen/captureScreenRect
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}
		// var pr interface{} = -5
		// v1p := 0

		// if instrT.ParamLen > 0 {
		pr := instrT.Params[0]
		v1p := 1
		// }

		var v1, v2, v3, v4 int

		var errT error

		if instrT.ParamLen > 1 {
			v1 = tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]), 0)
		} else {
			v1 = 0
		}

		if instrT.ParamLen > 2 {
			v2 = tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]), 0)
		} else {
			v2 = 0
		}

		if instrT.ParamLen > 3 {
			v3 = tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+2]), 0)
		} else {
			v3 = 0
		}

		if instrT.ParamLen > 4 {
			v4 = tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+3]), 0)
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
			p.SetVar(r, pr, errT)
			return ""
		}

		p.SetVar(r, pr, imageT)
		return ""

	case 50001: // genToken
		if instrT.ParamLen < 4 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
		v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))

		p.SetVar(r, pr, tk.GenerateToken(v1, v2, v3, p.ParamsToStrs(r, instrT, 4)...))

		return ""

	case 50003: // checkToken
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]

		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		p.SetVar(r, pr, tk.CheckToken(v1, p.ParamsToStrs(r, instrT, 2)...))

		return ""

	case 70001: // leClear
		leClear()

		return ""

	case 70003: // leLoadStr/leSetAll
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			// pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		leLoadString(v1)

		// p.SetVar(r, pr, rs)

		return ""

	case 70007: // leSaveStr/leGetAll
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		// v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		// leLoadString(v1)

		p.SetVar(r, pr, leSaveString())

		return ""

	case 70011: // leLoad/leLoadFile
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rs := leLoadFile(v1)

		p.SetVar(r, pr, rs)

		return ""

	case 70012: // leLoadClip
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		// v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rs := leLoadClip()

		p.SetVar(r, pr, rs)

		return ""

	case 70013: // leAppendFile
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rs := leAppendFile(v1)

		p.SetVar(r, pr, rs)

		return ""

	case 70015: // leLoadSSH
		if instrT.ParamLen < 5 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		pa := p.ParamsToStrs(r, instrT, v1p)

		var v1, v2, v3, v4, v5 string

		v1 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Host")
		v2 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Port")
		v3 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "User")
		v4 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Password")
		v5 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Path")

		v1 = p.GetSwitchVarValue(r, pa, "-host=", v1)     // tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 = p.GetSwitchVarValue(r, pa, "-port=", v2)     // tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
		v3 = p.GetSwitchVarValue(r, pa, "-user=", v3)     // tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))
		v4 = p.GetSwitchVarValue(r, pa, "-password=", v4) // tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+3]))
		if strings.HasPrefix(v4, "740404") {
			v4 = tk.DecryptStringByTXDEF(v4)
		}
		v5 = p.GetSwitchVarValue(r, pa, "-path=", v5) // tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+4]))

		sshT, errT := tk.NewSSHClient(v1, v2, v3, v4)

		if errT != nil {
			p.SetVar(r, pr, errT)
			if !leSilentG {
				tk.Pl("连接服务器失败：%v", errT)
			}

			return ""
		}

		defer sshT.Close()

		basePathT, errT := tk.EnsureBasePath("xie")
		if errT != nil {
			p.SetVar(r, pr, errT)
			if !leSilentG {
				tk.Pl("谢语言根路径不存在")
			}
			return ""
		}

		tmpFileT := filepath.Join(basePathT, "leSSHTmp.txt")

		errT = sshT.Download(v5, tmpFileT)

		if errT != nil {
			p.SetVar(r, pr, errT)
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
			p.SetVar(r, pr, errT)
			if !leSilentG {
				tk.Pl("加载文件失败（%v）：%v", tmpFileT, errT)
			}
			return ""
		}

		p.SetVar(r, pr, errT)

		return ""

	case 70017: // leSave/leSaveFile
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rs := leSaveFile(v1)

		p.SetVar(r, pr, rs)

		return ""

	case 70023: // leSaveClip
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		rs := leSaveClip()

		p.SetVar(r, pr, rs)

		return ""
	case 70025: // leSaveSSH
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		pa := p.ParamsToStrs(r, instrT, v1p)

		var v1, v2, v3, v4, v5 string

		v1 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Host")
		v2 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Port")
		v3 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "User")
		v4 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Password")
		v5 = tk.SafelyGetStringForKeyWithDefault(leSSHInfoG, "Path")

		v1 = p.GetSwitchVarValue(r, pa, "-host=", v1)     // tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 = p.GetSwitchVarValue(r, pa, "-port=", v2)     // tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
		v3 = p.GetSwitchVarValue(r, pa, "-user=", v3)     // tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))
		v4 = p.GetSwitchVarValue(r, pa, "-password=", v4) // tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+3]))
		if strings.HasPrefix(v4, "740404") {
			v4 = tk.DecryptStringByTXDEF(v4)
		}
		v5 = p.GetSwitchVarValue(r, pa, "-path=", v5) // tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+4]))

		// tk.Plvsr(v1, v2, v3, v4, v5)

		sshT, errT := tk.NewSSHClient(v1, v2, v3, v4)

		if errT != nil {
			p.SetVar(r, pr, errT)
			if !leSilentG {
				tk.Pl("连接服务器失败：%v", errT)
			}
			return ""
		}

		defer sshT.Close()

		basePathT, errT := tk.EnsureBasePath("xie")
		if errT != nil {
			p.SetVar(r, pr, errT)
			if !leSilentG {
				tk.Pl("谢语言根路径不存在")
			}
			return ""
		}

		tmpFileT := filepath.Join(basePathT, "leSSHTmp.txt")

		errT = leSaveFile(tmpFileT)
		if errT != nil {
			p.SetVar(r, pr, errT)
			if !leSilentG {
				tk.Pl("保存临时文件失败：%v", errT)
			}
			return ""
		}

		errT = sshT.Upload(tmpFileT, v5, pa...)

		if errT != nil {
			p.SetVar(r, pr, errT)
			if !leSilentG {
				tk.Pl("保存文件到服务器失败（%v）：%v", v5, errT)
			}

			return ""
		}

		p.SetVar(r, pr, errT)

		return ""

	case 70016: // leLoadUrl
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rs := leLoadUrl(v1)

		p.SetVar(r, pr, rs)

		return ""

	case 70027: // leInsert
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		rs := leInsertLine(v1, v2)

		p.SetVar(r, pr, rs)

		return ""

	case 70029: // leAppend/leAppendLine
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		rs := leAppendLine(v1)

		p.SetVar(r, pr, rs)

		return ""

	case 70033: // leSet/leSetLine
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		rs := leSetLine(v1, v2)

		p.SetVar(r, pr, rs)

		return ""

	case 70037: // leSetLines
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))
		v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))

		rs := leSetLines(v1, v2, v3)

		p.SetVar(r, pr, rs)

		return ""

	case 70039: // leRemove/leRemoveLine
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]))

		rs := leRemoveLine(v1)

		p.SetVar(r, pr, rs)

		return ""

	case 70043: // leRemoveLines
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p+1]))

		rs := leRemoveLines(v1, v2)

		p.SetVar(r, pr, rs)

		return ""

	case 70045: // leViewAll
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			v1p = 1
		}

		vs := p.ParamsToStrs(r, instrT, v1p)

		rs := leViewAll(vs...)

		if tk.IsError(rs) {
			if !leSilentG {
				tk.Pl("内部行编辑器操作失败：%v", rs)
			}
		}

		p.SetVar(r, pr, rs)

		return ""

	case 70047: // leView/leViewLine
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToInt(p.GetVarValue(r, instrT.Params[v1p]))

		rs := leViewLine(v1)

		if tk.IsError(rs) {
			if !leSilentG {
				tk.Pl("内部行编辑器操作失败：%v", rs)
			}
		}

		p.SetVar(r, pr, rs)

		return ""

	case 70049: // leSort
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToBool(p.GetVarValue(r, instrT.Params[v1p]))

		rs := leSort(v1)

		p.SetVar(r, pr, rs)

		return ""

	case 70051: // leEnc
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		vs := p.ParamsToStrs(r, instrT, v1p)

		rs := leConvertToUTF8(vs...)

		p.SetVar(r, pr, rs)

		return ""

	case 70061: // leLineEnd
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		pr := instrT.Params[0]
		v1p := 1

		if instrT.ParamLen > 1 {
			v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
			rs := leLineEnd(v1)
			p.SetVar(r, pr, rs)

			return ""
		}

		rs := leLineEnd()

		p.SetVar(r, pr, rs)

		return ""

	case 70071: // leSilent

		pr := instrT.Params[0]
		v1p := 1

		if instrT.ParamLen < 2 {
			p.SetVar(r, pr, leSilent())

			return ""
		}

		rs := leSilent(tk.ToBool(p.GetVarValue(r, instrT.Params[v1p])))

		p.SetVar(r, pr, rs)

		return ""

	case 70081: // leFind
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		rs := leFind(tk.ToStr(p.GetVarValue(r, instrT.Params[v1p])))

		if rs != nil {
			if !leSilentG {
				for _, v := range rs {
					tk.Pl("%v", v)
				}
			}
		}

		p.SetVar(r, pr, rs)

		return ""

	case 70083: // leReplace
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		rs := leReplace(v1, v2)

		if rs != nil {
			if !leSilentG {
				for _, v := range rs {
					tk.Pl("%v", v)
				}
				tk.Pl("共替换 %v 处", len(rs))
			}
		}

		p.SetVar(r, pr, rs)

		return ""

	case 70091: // leSSHInfo/leSshInfo
		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		p.SetVar(r, pr, leSSHInfoG)

		if !leSilentG {
			tk.Pl("leSSHInfo: %v", leSSHInfoG)
		}

		return ""

	case 70098: // leRun
		// if instrT.ParamLen < 2 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		// var pr interface{} = -5
		// v1p := 0

		// if instrT.ParamLen > 1 {
		pr := instrT.Params[0]
		v1p := 1
		// }

		codeT := leSaveString()

		inputT := p.GetVarValue(r, GetVarRefInParams(instrT.Params, v1p))

		if tk.IsUndefined(inputT) {
			inputT = nil
		}

		objT := p.GetVarValue(r, GetVarRefInParams(instrT.Params, v1p+1))

		obj1, ok := objT.(map[string]interface{})

		if !ok {
			obj1 = nil
		}

		vs := p.ParamsToStrs(r, instrT, v1p+2)

		rs := RunCode(codeT, inputT, obj1, vs...)

		p.SetVar(r, pr, rs)

		p.SetVar(r, pr, rs)

		return ""
	case 80001: // getMimeType
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		extT := filepath.Ext(v1)

		p.SetVar(r, pr, tk.GetMimeTypeByExt(extT))

		return ""

	case 90101: // archiveFilesToZip
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))
		vs := p.ParamsToList(r, instrT, v1p+1)

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
			CompressionLevel:  flate.DefaultCompression,
			OverwriteExisting: tk.IfSwitchExistsWhole(args1T, "-overwrite"),
			MkdirAll:          tk.IfSwitchExistsWhole(args1T, "-makeDirs"),
			// SelectiveCompression:   true,
			// ImplicitTopLevelFolder: false,
			// ContinueOnError:        false,
			FileMethod: archiver.Deflate,
		}

		errT := z.Archive(fileNamesT, v1)
		// if errT != nil {
		// 	tk.Plv(errT)
		// 	// tk.AppendStringToFile(tk.Spr("Archive error(%v): %v", filePathA, errT), logFileNameT)
		// 	tk.SaveStringToFile(tk.Spr("Archive error(%v): %v", filePathA, errT), errFileNameT)
		// 	return
		// }

		p.SetVar(r, pr, errT)

		return ""

	case 90111: // extractFilesFromZip
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		destT := "."

		if instrT.ParamLen > 2 {
			destT = tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))
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

		p.SetVar(r, pr, errT)

		return ""

	// case 90201: // compressData
	// 	if instrT.ParamLen < 2 {
	// 		return p.Errf(r, "not enough parameters(参数不够)")
	// 	}

	// 	pr := instrT.Params[0]
	// 	v1p := 1

	// 	v1 := p.GetVarValue(r, instrT.Params[v1p])

	// 	v1vb, ok := v1.([]byte)

	// 	if !ok {
	// 		v1vs, ok := v1.(string)

	// 		if ok {
	// 			v1vb = []byte(v1vs)
	// 		} else {
	// 			return p.Errf(r, "参数格式错误（invalid param type）：%T(%v)", v1, v1)
	// 		}
	// 	}

	// 	vs := p.ParamsToStrs(r, instrT, v1p+1)

	// 	methodT := p.GetSwitchVarValue(r, vs, "-method=", "gzip")

	// 	nameT := p.GetSwitchVarValue(r, vs, "-name=", "")
	// 	commentT := p.GetSwitchVarValue(r, vs, "-comment=", "")
	// 	timeT := tk.ToTime(p.GetSwitchVarValue(r, vs, "-time=", ""), time.Now()).(time.Time)

	// 	var buf bytes.Buffer
	// 	zw := gzip.NewWriter(&buf)

	// 	if methodT == "gzip" {
	// 		zw.Name = nameT
	// 		zw.Comment = commentT
	// 		zw.ModTime = timeT

	// 		_, err := zw.Write(v1vb)
	// 		if err != nil {
	// 			p.SetVar(r, pr, fmt.Errorf("压缩数据时发生错误（failed to compress data）：%v", err))
	// 			return ""
	// 		}

	// 		if err := zw.Close(); err != nil {
	// 			p.SetVar(r, pr, fmt.Errorf("压缩数据关闭文件时发生错误（failed to compress data file）：%v", err))
	// 			return ""
	// 		}

	// 		p.SetVar(r, pr, buf.Bytes())
	// 		return ""
	// 	}

	// 	p.SetVar(r, pr, fmt.Errorf("不支持的压缩格式（unsupported compress method）：%v", methodT))
	// 	return ""

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
					return p.Errf(r, "初始化WEB图形界面环境失败")
				}

				rs = tk.DownloadFile("http://xie.topget.org/pub/scapp.exe", applicationPathT, "scapp.exe")

				if tk.IsErrorString(rs) {
					return p.Errf(r, "初始化WEB图形界面环境失败")
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
				return p.Errf(r, "更新WEB图形界面环境失败")
			}

			rs = tk.DownloadFile("http://xie.topget.org/pub/scapp.exe", applicationPathT, "scapp.exe")

			if tk.IsErrorString(rs) {
				return p.Errf(r, "初始化WEB图形界面环境失败")
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
					return p.Errf(r, "初始化WEB图形界面环境失败")
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
					return p.Errf(r, "解压缩图形环境压缩包失败：%v", errT)
				}

			}
		}

		return ""

	case 200001: // sshConnect/sshOpen
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		pa := p.ParamsToStrs(r, instrT, v1p)

		var v1, v2, v3, v4 string

		v1 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-host=", v1))
		v2 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-port=", "22"))
		v3 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-user=", v3))
		v4 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-password=", v4))
		if strings.HasPrefix(v4, "740404") {
			v4 = strings.TrimSpace(tk.DecryptStringByTXDEF(v4))
		}

		if v1 == "" {
			p.SetVar(r, pr, fmt.Errorf("host不能为空"))
			return ""
		}

		if v2 == "" {
			p.SetVar(r, pr, fmt.Errorf("port不能为空"))
			return ""
		}

		if v3 == "" {
			p.SetVar(r, pr, fmt.Errorf("user不能为空"))
			return ""
		}

		if v4 == "" {
			p.SetVar(r, pr, fmt.Errorf("password不能为空"))
			return ""
		}

		sshT, errT := tk.NewSSHClient(v1, v2, v3, v4)

		if errT != nil {
			p.SetVar(r, pr, errT)

			return ""
		}

		p.SetVar(r, pr, sshT)
		return ""

	case 200003: // sshClose
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		sshT, ok := v1.(*goph.Client)

		if !ok {
			return p.Errf(r, "参数类型错误：%T(%v)", sshT, sshT)
		}

		rsT := sshT.Close()

		p.SetVar(r, pr, rsT)
		return ""

	case 200011: // sshUpload
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		pa := p.ParamsToStrs(r, instrT, v1p)

		var v1, v2, v3, v4, v5, v6 string

		v1 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-host=", v1))
		v2 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-port=", v2))
		v3 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-user=", v3))
		v4 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-password=", v4))
		if strings.HasPrefix(v4, "740404") {
			v4 = strings.TrimSpace(tk.DecryptStringByTXDEF(v4))
		}
		v5 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-path=", v5))
		v6 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-remotePath=", v6))

		withProgressT := tk.IfSwitchExistsWhole(pa, "-progress")

		if v1 == "" {
			p.SetVar(r, pr, fmt.Errorf("host不能为空"))
			return ""
		}

		if v2 == "" {
			p.SetVar(r, pr, fmt.Errorf("port不能为空"))
			return ""
		}

		if v3 == "" {
			p.SetVar(r, pr, fmt.Errorf("user不能为空"))
			return ""
		}

		if v4 == "" {
			p.SetVar(r, pr, fmt.Errorf("password不能为空"))
			return ""
		}

		if v5 == "" {
			p.SetVar(r, pr, fmt.Errorf("path不能为空"))
			return ""
		}

		if v6 == "" {
			p.SetVar(r, pr, fmt.Errorf("remotePath不能为空"))
			return ""
		}

		sshT, errT := tk.NewSSHClient(v1, v2, v3, v4)

		if errT != nil {
			p.SetVar(r, pr, errT)

			return ""
		}

		defer sshT.Close()

		if withProgressT {
			fmt.Println()
			errT = sshT.UploadWithProgressFunc(v5, v6, func(i interface{}) interface{} {
				fmt.Printf("\rprogress: %v                ", tk.IntToKMGT(i))
				return ""
			}, pa...)
		} else {
			errT = sshT.Upload(v5, v6, pa...)
		}

		if errT != nil {
			p.SetVar(r, pr, errT)

			return ""
		}

		p.SetVar(r, pr, nil)
		return ""

	case 200013: // sshUploadBytes
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 2

		v0 := p.GetVarValue(r, instrT.Params[1])

		// tk.Plvx(v0)

		v0v, ok := v0.([]byte)

		if !ok {
			v0vs, ok := v0.(string)

			if !ok {
				return p.Errf(r, "参数类型错误：%T(%v)", v0, v0)
			}

			v0v = []byte(v0vs)
		}

		pa := p.ParamsToStrs(r, instrT, v1p)

		var v1, v2, v3, v4, v6 string

		v2 = "22"

		v1 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-host=", v1))
		v2 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-port=", v2))
		v3 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-user=", v3))
		v4 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-password=", v4))
		if strings.HasPrefix(v4, "740404") {
			v4 = strings.TrimSpace(tk.DecryptStringByTXDEF(v4))
		}
		// v5 = strings.TrimSpace(p.GetSwitchVarValue(pa, "-path=", v5))
		v6 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-remotePath=", v6))

		if v1 == "" {
			p.SetVar(r, pr, fmt.Errorf("host不能为空"))
			return ""
		}

		if v2 == "" {
			p.SetVar(r, pr, fmt.Errorf("port不能为空"))
			return ""
		}

		if v3 == "" {
			p.SetVar(r, pr, fmt.Errorf("user不能为空"))
			return ""
		}

		if v4 == "" {
			p.SetVar(r, pr, fmt.Errorf("password不能为空"))
			return ""
		}

		// if v5 == "" {
		// 	p.SetVar(r, pr, fmt.Errorf("path不能为空"))
		// 	return ""
		// }

		if v6 == "" {
			p.SetVar(r, pr, fmt.Errorf("remotePath不能为空"))
			return ""
		}

		sshT, errT := tk.NewSSHClient(v1, v2, v3, v4)

		if errT != nil {
			p.SetVar(r, pr, errT)

			return ""
		}

		defer sshT.Close()

		errT = sshT.UploadFileContent(v0v, v6, pa...)

		if errT != nil {
			p.SetVar(r, pr, errT)

			return ""
		}

		p.SetVar(r, pr, nil)
		return ""

	case 200021: // sshDownload
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		pa := p.ParamsToStrs(r, instrT, v1p)

		var v1, v2, v3, v4, v5, v6 string

		v2 = "22"

		v1 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-host=", v1))
		v2 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-port=", v2))
		v3 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-user=", v3))
		v4 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-password=", v4))
		if strings.HasPrefix(v4, "740404") {
			v4 = strings.TrimSpace(tk.DecryptStringByTXDEF(v4))
		}
		v5 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-path=", v5))
		v6 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-remotePath=", v6))

		if v1 == "" {
			p.SetVar(r, pr, fmt.Errorf("host不能为空"))
			return ""
		}

		if v2 == "" {
			p.SetVar(r, pr, fmt.Errorf("port不能为空"))
			return ""
		}

		if v3 == "" {
			p.SetVar(r, pr, fmt.Errorf("user不能为空"))
			return ""
		}

		if v4 == "" {
			p.SetVar(r, pr, fmt.Errorf("password不能为空"))
			return ""
		}

		if v5 == "" {
			p.SetVar(r, pr, fmt.Errorf("path不能为空"))
			return ""
		}

		if v6 == "" {
			p.SetVar(r, pr, fmt.Errorf("remotePath不能为空"))
			return ""
		}

		sshT, errT := tk.NewSSHClient(v1, v2, v3, v4)

		if errT != nil {
			p.SetVar(r, pr, errT)

			return ""
		}

		defer sshT.Close()

		errT = sshT.Download(v6, v5)

		if errT != nil {
			p.SetVar(r, pr, errT)

			return ""
		}

		p.SetVar(r, pr, nil)
		return ""

	case 200023: // sshDownloadBytes
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		pr := instrT.Params[0]
		v1p := 1

		pa := p.ParamsToStrs(r, instrT, v1p)

		var v1, v2, v3, v4, v6 string

		v2 = "22"

		v1 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-host=", v1))
		v2 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-port=", v2))
		v3 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-user=", v3))
		v4 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-password=", v4))
		if strings.HasPrefix(v4, "740404") {
			v4 = strings.TrimSpace(tk.DecryptStringByTXDEF(v4))
		}
		// v5 = strings.TrimSpace(p.GetSwitchVarValue(pa, "-path=", v5))
		v6 = strings.TrimSpace(p.GetSwitchVarValue(r, pa, "-remotePath=", v6))

		if v1 == "" {
			p.SetVar(r, pr, fmt.Errorf("host不能为空"))
			return ""
		}

		if v2 == "" {
			p.SetVar(r, pr, fmt.Errorf("port不能为空"))
			return ""
		}

		if v3 == "" {
			p.SetVar(r, pr, fmt.Errorf("user不能为空"))
			return ""
		}

		if v4 == "" {
			p.SetVar(r, pr, fmt.Errorf("password不能为空"))
			return ""
		}

		// if v5 == "" {
		// 	p.SetVar(r, pr, fmt.Errorf("path不能为空"))
		// 	return ""
		// }

		if v6 == "" {
			p.SetVar(r, pr, fmt.Errorf("remotePath不能为空"))
			return ""
		}

		sshT, errT := tk.NewSSHClient(v1, v2, v3, v4)

		if errT != nil {
			p.SetVar(r, pr, errT)

			return ""
		}

		defer sshT.Close()

		rsT, errT := sshT.GetFileContent(v6)

		if errT != nil {
			p.SetVar(r, pr, errT)

			return ""
		}

		p.SetVar(r, pr, rsT)
		return ""

	case 210001: // excelNew
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 0 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		// v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		f := excelize.NewFile()
		if f == nil {
			p.SetVar(r, pr, fmt.Errorf("无法创建新的Excel文件（failed to create Excel file）"))
			return ""
		}

		p.SetVar(r, pr, f)
		return ""

	case 210003: // excelOpen
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p]))

		f, err := excelize.OpenFile(v1)
		if err != nil {
			p.SetVar(r, pr, err)
			return ""
		}

		p.SetVar(r, pr, f)
		return ""

	case 210005: // excelClose
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		f1, ok := v1.(*excelize.File)

		if !ok {
			return p.Errf(r, "参数类型错误：%T(%v)", v1, v1)
		}

		// tk.Pl("close excel: %v", f1)

		err := f1.Close()
		if err != nil {
			p.SetVar(r, pr, err)
			return ""
		}

		p.SetVar(r, pr, nil)
		return ""

	case 210007: // excelSaveAs
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		f1, ok := v1.(*excelize.File)

		if !ok {
			return p.Errf(r, "参数类型错误：%T(%v)", v1, v1)
		}

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		err := f1.SaveAs(v2)
		if err != nil {
			p.SetVar(r, pr, err)
			return ""
		}

		p.SetVar(r, pr, nil)
		return ""

	case 210009: // excelWrite
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		f1, ok := v1.(*excelize.File)

		if !ok {
			return p.Errf(r, "参数类型错误：%T(%v)", v1, v1)
		}

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		err := f1.Write(v2.(io.Writer))
		if err != nil {
			p.SetVar(r, pr, err)
			return ""
		}

		p.SetVar(r, pr, nil)
		return ""

	case 210101: // excelReadSheet
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		f1, ok := v1.(*excelize.File)

		if !ok {
			return p.Errf(r, "参数类型错误：%T(%v)", v1, v1)
		}

		v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		v2i, ok := v2.(int)

		if ok {
			rowsT, errT := f1.GetRows(f1.GetSheetName(v2i))

			if errT != nil {
				p.SetVar(r, pr, errT)
				return ""
			}

			p.SetVar(r, pr, rowsT)
			return ""
		}

		rowsT, errT := f1.GetRows(tk.ToStr(v2))
		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}
		p.SetVar(r, pr, rowsT)
		return ""

	case 210103: // excelReadCell
		if instrT.ParamLen < 3 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 3 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		f1, ok := v1.(*excelize.File)

		if !ok {
			return p.Errf(r, "参数类型错误：%T(%v)", v1, v1)
		}

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))

		valueT, errT := f1.GetCellValue(v2, v3)
		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		p.SetVar(r, pr, valueT)
		return ""

	case 210105: // excelWriteCell/excelSetCell
		if instrT.ParamLen < 4 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 4 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		// tk.Pl("v1p: %v, %#v", v1p, v1)

		f1, ok := v1.(*excelize.File)

		if !ok {
			return p.Errf(r, "参数类型错误：%T(%v)", v1, v1)
		}

		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		v3 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+2]))

		v4 := p.GetVarValue(r, instrT.Params[v1p+3])

		errT := f1.SetCellValue(v2, v3, v4)
		if errT != nil {
			p.SetVar(r, pr, errT)
			return ""
		}

		p.SetVar(r, pr, nil)
		return ""

	case 210201: // excelGetSheetList
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		f1, ok := v1.(*excelize.File)

		if !ok {
			return p.Errf(r, "参数类型错误：%T(%v)", v1, v1)
		}

		p.SetVar(r, pr, f1.GetSheetList())
		return ""

	case 300101: // awsSign
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		var pr interface{} = -5
		v1p := 0

		if instrT.ParamLen > 2 {
			pr = instrT.Params[0]
			v1p = 1
		}

		v1 := p.GetVarValue(r, instrT.Params[v1p])
		v2 := tk.ToStr(p.GetVarValue(r, instrT.Params[v1p+1]))

		var signT string

		nv1, ok := v1.(url.Values)

		if ok {
			signT = awsapi.Sign(nv1, v2)
			p.SetVar(r, pr, signT)
			return ""
		}

		nv2, ok := v1.(map[string]interface{})
		if ok {
			signT = awsapi.Sign(tk.MapToPostDataI(nv2), v2)
			p.SetVar(r, pr, signT)
			return ""

		}

		nv3, ok := v1.(map[string]string)
		if ok {
			signT = awsapi.Sign(tk.MapToPostData(nv3), v2)
			p.SetVar(r, pr, signT)
			return ""

		}

		return p.Errf(r, "failed to sign AWS, unsupport type: %T(%v)", v1, v1)

	case 400000: // guiInit
		// if instrT.ParamLen < 1 {
		// 	return p.Errf(r, "not enough parameters(参数不够)")
		// }

		var pr interface{} = -5
		// v1p := 0

		if instrT.ParamLen > 1 {
			pr = instrT.Params[0]
			// v1p = 1
		}

		v0, ok := p.GetVarValue(r, ParseVar("$guiG")).(tk.TXDelegate)

		if !ok {
			return p.Errf(r, "全局变量guiG不存在（$guiG not exists）")
		}

		rs := v0("init", p, nil)

		// if tk.IsErrX(rs) {
		p.SetVar(r, pr, rs)
		// 	return p.ErrStrf(tk.GetErrStrX(rs))
		// }

		return ""

	case 400001: // alert
		if instrT.ParamLen < 1 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		v0, ok := p.GetVarValue(r, ParseVar("$guiG")).(tk.TXDelegate)

		if !ok {
			return p.Errf(r, "global variable $guiG not exists")
		}

		v1p := 0

		v1 := p.GetVarValue(r, instrT.Params[v1p])

		rs := v0("showInfo", p, nil, "", fmt.Sprintf("%v", v1))
		tk.Plv(rs)

		if tk.IsErrX(rs) {
			return p.Errf(r, tk.GetErrStrX(rs))
		}

		return ""

	case 400003: // msgBox/showInfo
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// var pr interface{} = -5
		// v1p := 0

		// if instrT.ParamLen > 1 {
		// 	pr = instrT.Params[0]
		// 	v1p = 1
		// }

		v0, ok := p.GetVarValue(r, ParseVar("$guiG")).(tk.TXDelegate)

		if !ok {
			return p.Errf(r, "全局变量guiG不存在（$guiG not exists）")
		}

		v1p := 0

		// v1 := p.GetVarValue(r, instrT.Params[v1p])
		// v2 := p.GetVarValue(r, instrT.Params[v1p+1])
		vs := p.ParamsToList(r, instrT, v1p)

		rs := v0("showInfo", p, nil, vs...)

		if tk.IsErrX(rs) {
			return p.Errf(r, tk.GetErrStrX(rs))
		}

		return ""

	case 400005: // showError
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// var pr interface{} = -5
		// v1p := 0

		// if instrT.ParamLen > 1 {
		// 	pr = instrT.Params[0]
		// 	v1p = 1
		// }

		v0, ok := p.GetVarValue(r, ParseVar("$guiG")).(tk.TXDelegate)

		if !ok {
			return p.Errf(r, "全局变量guiG不存在（$guiG not exists）")
		}

		v1p := 0

		// v1 := p.GetVarValue(r, instrT.Params[v1p])
		// v2 := p.GetVarValue(r, instrT.Params[v1p+1])
		vs := p.ParamsToList(r, instrT, v1p)

		rs := v0("showError", p, nil, vs...)

		if tk.IsErrX(rs) {
			return p.Errf(r, tk.GetErrStrX(rs))
		}

		return ""

	case 400011: // getConfirm
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough parameters(参数不够)")
		}

		// var pr interface{} = -5
		// v1p := 0

		// if instrT.ParamLen > 1 {
		pr := instrT.Params[0]
		v1p := 1
		// }

		v0, ok := p.GetVarValue(r, ParseVar("$guiG")).(tk.TXDelegate)

		if !ok {
			return p.Errf(r, "全局变量guiG不存在（$guiG not exists）")
		}

		// v1p := 0

		// v1 := p.GetVarValue(r, instrT.Params[v1p])
		// v2 := p.GetVarValue(r, instrT.Params[v1p+1])

		vs := p.ParamsToList(r, instrT, v1p)

		rs := v0("getConfirm", p, nil, vs...)

		if tk.IsErrX(rs) {
			return p.Errf(r, tk.GetErrStrX(rs))
		}

		p.SetVar(r, pr, rs)
		return ""

	case 400031: // guiNewWindow
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough paramters")
		}

		pr := instrT.Params[0]
		v1p := 1

		v0, ok := p.GetVar(r, "guiG").(tk.TXDelegate)

		if !ok {
			return p.Errf(r, "$guiG not exists")
		}

		vs := p.ParamsToList(r, instrT, v1p)

		rs := v0("newWindow", p, instrT, vs...)

		p.SetVar(r, pr, rs)
		return ""
	case 410001: // guiMethod/guiMt
		if instrT.ParamLen < 2 {
			return p.Errf(r, "not enough paramters")
		}

		pr := instrT.Params[0]
		v1p := 1

		v0, ok := p.GetVarValue(r, ParseVar("$guiG")).(tk.TXDelegate)

		if !ok {
			return p.Errf(r, "$guiG not exists")
		}

		vs := p.ParamsToList(r, instrT, v1p)

		rs := v0("method", p, instrT, vs...)

		p.SetVar(r, pr, rs)
		return ""

		// end of switch

		//// ------ cmd end

	}

	return p.Errf(r, "unknown instr: %v", instrT)
}

func (p *XieVM) Run(posA ...int) interface{} {
	// tk.Pl("%#v", p)
	p.Running.CodePointer = 0
	if len(posA) > 0 {
		p.Running.CodePointer = posA[0]
	}

	if len(p.Running.CodeList) < 1 {
		return tk.Undefined
	}

	for {
		if GlobalsG.VerboseLevel > 1 {
			tk.Pl("-- RunInstr [%v] %v", p.Running.CodePointer, tk.LimitString(p.Running.Source[p.Running.CodeSourceMap[p.Running.CodePointer]], 50))
		}

		resultT := RunInstr(p, p.Running, &p.Running.InstrList[p.Running.CodePointer])

		c1T, ok := resultT.(int)

		if ok {
			p.Running.CodePointer = c1T
		} else {
			if tk.IsError(resultT) {
				if p.Running.ErrorHandler > -1 {
					p.SetVarGlobal("lastLineG", p.Running.CodeSourceMap[p.Running.CodePointer]+1)
					p.SetVarGlobal("errorMessageG", "runtime error")
					p.SetVarGlobal("errorDetailG", tk.GetErrStrX(resultT))
					// p.Stack.Push(tk.GetErrStrX(resultT))
					// p.Stack.Push("runtime error")
					// p.Stack.Push(p.Running.CodeSourceMap[p.Running.CodePointer] + 1)

					p.Running.CodePointer = p.Running.ErrorHandler

					continue
				}
				// tk.Plo(1.2, p.Running, p.RootFunc)
				p.RunDeferUpToRoot(p.Running)
				return p.Errf(p.Running, "[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStrX(resultT))
				// tk.Pl("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), p.CodeSourceMapM[p.CodePointerM]+1, tk.GetErrStr(rs))
				// break
			}

			rs, ok := resultT.(string)

			if !ok {
				p.RunDeferUpToRoot(p.Running)
				return p.Errf(p.Running, "return result error: (%T)%v", resultT, resultT)
			}

			if tk.IsErrStrX(rs) {
				p.RunDeferUpToRoot(p.Running)
				return p.Errf(p.Running, "[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStr(rs))
				// tk.Pl("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), p.CodeSourceMapM[p.CodePointerM]+1, tk.GetErrStr(rs))
				// break
			}

			if rs == "" {
				p.Running.CodePointer++

				if p.Running.CodePointer >= len(p.Running.CodeList) {
					break
				}
			} else if rs == "exit" {
				break
				// } else if rs == "cont" {
				// 	return p.Errf(r, "无效指令: %v", rs)
				// } else if rs == "brk" {
				// 	return p.Errf(r, "无效指令: %v", rs)
			} else {
				tmpI := tk.StrToInt(rs)

				if tmpI < 0 {
					p.RunDeferUpToRoot(p.Running)

					return p.Errf(p.Running, "invalid instr: %v", rs)
				}

				if tmpI >= len(p.Running.CodeList) {
					p.RunDeferUpToRoot(p.Running)
					return p.Errf(p.Running, "instr index out of range: %v(%v)/%v", tmpI, rs, len(p.Running.CodeList))
				}

				p.Running.CodePointer = tmpI
			}

		}

	}

	rsi := p.RunDeferUpToRoot(p.Running)

	if tk.IsErrX(rsi) {
		return tk.ErrStrf("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStrX(rsi))
	}

	// tk.Pl(tk.ToJSONX(p, "-indent", "-sort"))

	outT, ok := p.GetFuncContext(p.Running, 0).Vars["outG"]
	if !ok {
		return tk.Undefined
	}

	return outT

}

// 运行一段代码，可以是代码或编译后，r是运行上下文，为nil会新建一个，会进行函数压栈（以便新的运行上下文中可以deferUpToRoot），入参通过inputA传入后从$inputL访问，出参通过$outL传出
func RunCodePiece(p *XieVM, r interface{}, codeA interface{}, inputA interface{}, runDeferA bool) (rst interface{}) {
	defer func() {
		if r1 := recover(); r1 != nil {
			rst = fmt.Errorf("runtime exception: %v\n%v", r1, string(debug.Stack()))

			return
		}
	}()

	var rp *RunningContext

	if r == nil {
		r = NewRunningContext()
		if tk.IsError(r) {
			return p.Errf(rp, "[%v](xie) runtime error, failed to initialize running contexte: %v", tk.GetNowTimeStringFormal(), r)
		}
	}

	rp = r.(*RunningContext)

	errT := rp.Load(codeA)

	if errT != nil {
		return p.Errf(rp, "[%v](xie) runtime error, failed to load code: %v", tk.GetNowTimeStringFormal(), errT)
	}

	if len(rp.CodeList) < 1 {
		return tk.Undefined
	}

	rp.CodePointer = 0

	rp.PushFunc()

	func1 := rp.FuncStack.Peek().(*FuncContext)
	func1.Vars["inputL"] = inputA

	for {
		// tk.Pl("-- [%v] %v", p.CodePointerM, tk.LimitString(p.SourceM[p.CodeSourceMapM[p.CodePointerM]], 50))
		resultT := RunInstr(p, rp, &rp.InstrList[rp.CodePointer])
		if GlobalsG.VerboseLevel > 1 {
			tk.Pl("--- RunInstr: %#v: %#v", rp.InstrList[rp.CodePointer], resultT)
		}

		c1T, ok := resultT.(int)

		if ok {
			rp.CodePointer = c1T
		} else {
			if tk.IsError(resultT) {
				if GlobalsG.VerboseLevel > 0 {
					tk.Pln(p.Errf(rp, "[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), resultT))
				}
				// tk.Pln("error: ", resultT)

				if runDeferA {
					rp.RunDeferUpToRoot(p)
				}
				return p.Errf(rp, "[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), resultT)
				// tk.Pl("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), p.CodeSourceMapM[p.CodePointerM]+1, tk.GetErrStr(rs))
				// break
			}

			rs, ok := resultT.(string)

			if !ok {
				if runDeferA {
					rp.RunDeferUpToRoot(p)
				}
				return p.Errf(rp, "return result error: (%T)%v", resultT, resultT)
			}

			// if tk.IsErrStrX(rs) {
			// 	rp.RunDeferUpToRoot(p)
			// 	return p.Errf(rp, "[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStr(rs))
			// 	// tk.Pl("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), p.CodeSourceMapM[p.CodePointerM]+1, tk.GetErrStr(rs))
			// 	// break
			// }

			if rs == "" {
				rp.CodePointer++

				if rp.CodePointer >= len(rp.CodeList) {
					break
				}
			} else if rs == "exit" {

				break
				// } else if rs == "cont" {
				// 	return p.Errf(r, "无效指令: %v", rs)
				// } else if rs == "brk" {
				// 	return p.Errf(r, "无效指令: %v", rs)
			} else {
				tmpI := tk.StrToInt(rs)

				if tmpI < 0 {
					if runDeferA {
						rp.RunDeferUpToRoot(p)
					}

					return p.Errf(rp, "invalid instr: %v", rs)
				}

				if tmpI >= len(rp.CodeList) {
					if runDeferA {
						rp.RunDeferUpToRoot(p)
					}
					return p.Errf(rp, "instr index out of range: %v(%v)/%v", tmpI, rs, len(rp.CodeList))
				}

				rp.CodePointer = tmpI
			}

		}

	}

	if runDeferA {
		rsi := rp.RunDeferUpToRoot(p)

		if tk.IsError(rsi) {
			return tk.ErrStrf("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), rsi)
		}
	}

	// tk.Pl(tk.ToJSONX(p, "-indent", "-sort"))
	// tk.Pl("%v", tk.ToJSONX(func1, "-indent", "-sort"))
	outT, ok := func1.Vars["outL"]
	rp.PopFunc()

	if !ok {
		return tk.Undefined
	}

	return outT
}

func RunCode(codeA interface{}, inputA interface{}, objA map[string]interface{}, optsA ...string) (rst interface{}) {
	defer func() {
		if r1 := recover(); r1 != nil {
			rst = fmt.Errorf("runtime exception: %v\n%v", r1, string(debug.Stack()))

			return
		}

		// rst = fmt.Errorf("runtime exception: unknown error")

		// return
	}()

	vmAnyT := NewVM()

	if tk.IsError(vmAnyT) {
		return vmAnyT
	}

	vmT := vmAnyT.(*XieVM)

	if len(optsA) > 0 {
		vmT.SetVar(vmT.Running, "argsG", optsA)
	}

	for k, v := range objA {
		vmT.SetVar(vmT.Running, k, v)
	}

	if inputA != nil {
		vmT.SetVar(vmT.Running, "inputG", inputA)
	}

	lrs := vmT.Load(vmT.Running, codeA)

	if tk.IsError(lrs) {
		return lrs
	}

	rs := vmT.Run()

	return rs
}

func (p *XieVM) Debug() {
	tk.Pln(tk.ToJSONX(p, "-indent", "-sort"))
}

type ExprElement struct {
	// 0: value, 1: operator， 5: eval, 6: (, 7: ), 9: end
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

func SplitExpr(strA string) ([]ExprElement, error) {
	runeT := []rune(strings.TrimSpace(strA))

	elementsT := make([]ExprElement, 0)

	// 0: start, 1: operator, 2: value, 3: value in quote such as "abc", 4: wait slash in quote such as "ab\n", 5: blank after value, 6: in {}
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
			case '{':
				stateT = 6
				valueT = ``
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
			case '}':
				return nil, fmt.Errorf("无法匹配的花括号")
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
			case '{':
				stateT = 6
				valueT = ``
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
			case '}':
				return nil, fmt.Errorf("无法匹配的花括号")
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
			case '{':
				stateT = 6
				opT = strings.TrimSpace(opT)
				if len(opT) > 0 {
					elementsT = append(elementsT, ExprElement{Type: 1, Priority: OperatorPriorityMap[opT], Value: opT})
				}
				opT = ""
				break
			case '}':
				return nil, fmt.Errorf("无法匹配的花括号")
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
		} else if stateT == 6 { // 值中的花括号内
			switch v {
			case '}':
				// valueT += `}`
				elementsT = append(elementsT, ExprElement{Type: 5, Priority: 0, Value: valueT})
				stateT = 5
				break
			default:
				valueT += string(v)
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
	} else if stateT == 6 {
		return nil, fmt.Errorf("表达式格式错误花引号不匹配")
	}

	// tk.Plv(elementsT)

	backElementsT := make([]ExprElement, 0)

	opStackT := tk.NewSimpleStack(len(elementsT) + 1)

	for _, v := range elementsT {
		// tk.Pl("process %v, %v, %v", v, backElementsT, opStackT)
		if v.Type == 0 {
			backElementsT = append(backElementsT, v)
		} else if v.Type == 5 {
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

func QuickEval(strA string, p *XieVM, r *RunningContext) interface{} {

	listT, errT := SplitExpr(strA)

	if errT != nil {
		return fmt.Errorf("分析表达式失败：%v", errT)
	}

	valueStackT := tk.NewSimpleStack(len(listT) + 1)

	for _, v := range listT {
		// tk.Pl("item: %v", v)
		if v.Type == 0 {
			v1T := ParseVar(v.Value)
			vv1T := p.GetVarValue(r, v1T)

			// tk.Plvx(vv1T)

			valueStackT.Push(vv1T)
		} else if v.Type == 5 { // eval

			// v1T := p.ParseVar(v.Value)
			// vv1T := p.EvalExpression(v1T)

			// tk.Pl("v.Value: %v", v.Value)

			codeStrT := strings.ReplaceAll(v.Value, "<BR />", "\n")

			compiledT := Compile(codeStrT)

			if tk.IsError(compiledT) {
				return fmt.Errorf("failed to compile the instructions: %v(%v)", compiledT, v.Value)
			}

			compiledObjT := compiledT.(*CompiledCode)

			for i1, i1v := range compiledObjT.InstrList {
				rs := RunInstr(p, r, &i1v)

				if tk.IsError(rs) {
					return fmt.Errorf("failed to run the instruction[%v]: %v(%v)", i1, rs, i1v)
				}
			}

			// 表达式应将结果存入$tmp
			valueStackT.Push(p.GetCurrentFuncContext(r).Tmp)
		} else if v.Type == 1 {
			switch v.Value {
			case "1-":
				v1 := valueStackT.Pop()

				vr := tk.GetNegativeResult(v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "1!":
				v1 := valueStackT.Pop()

				vr := tk.GetNegativeResult(v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "+":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				// tk.Pl("|%#v| |%#v|", v1, v2)

				vr := tk.GetAddResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "-":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetMinusResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "*":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetMultiplyResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "/":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetDivResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "%":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetModResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "==":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetEQResult(v2, v1)

				// tk.Plvsr(v1, v2, vr)
				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "!=", "<>":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetNEQResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case ">":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetGTResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "<":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetLTResult(v2, v1)

				// tk.Plo("<", v2, v1, vr)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case ">=":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetGETResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "<=":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetLETResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "&&":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetANDResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "||":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetORResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "&":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetBitANDResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "|":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetBitORResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "^":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetBitXORResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
				}

				valueStackT.Push(vr)
			case "&^":
				v1 := valueStackT.Pop()

				v2 := valueStackT.Pop()

				vr := tk.GetBitANDNOTResult(v2, v1)

				if tk.IsErrX(vr) {
					return fmt.Errorf("failed to cal the expression: %v", vr)
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

// func (p *XieVM) EvalExpressionNoGroup(strA string, valuesA *map[string]interface{}) interface{} {
// 	// strT := strA[1 : len(strA)-1]
// 	// tk.Pl("EvalExpressionNoGroup: %v", strA)
// 	// tk.Pl("tmpValues: %v", valuesA)
// 	if strings.HasPrefix(strA, "?") {
// 		instrT := p.NewInstr(strA[1:], valuesA)

// 		if instrT.Code == InstrNameSet["invalidInstr"] {
// 			return fmt.Errorf("指令分析失败：%v", instrT.Params[0].Value)
// 		}

// 		rsT := p.RunLine(0, instrT)

// 		nsv, ok := rsT.(string)

// 		if ok {
// 			if tk.IsErrStr(nsv) {
// 				return fmt.Errorf("计算失败：%v", tk.GetErrStr(nsv))
// 			}
// 		}

// 		// keyT := "~" + tk.IntToStr(len(valuesA))

// 		// valuesA[keyT] = p.Pop()

// 		strA = "$tmp"

// 	}

// 	listT := strings.SplitN(strA, " ", -1)

// 	// lenT := len(listT)

// 	// opListT := make([][]interface{}, 0, lenT)

// 	stateT := 0 // 0: initial, 1: first value ready, 2: operator ready, 3: second value ready

// 	opT := []interface{}{nil, nil, nil}

// 	// valuesT := make([]interface{})

// 	for _, v := range listT {
// 		v = strings.TrimSpace(v)

// 		if v == "" {
// 			continue
// 		}

// 		// tk.Pl("v: %v", v)

// 		if tk.InStrings(v, "+", "-", "*", "/", "%", "!", "&&", "||", "==", "!=", "<>", ">", "<", ">=", "<=", "&", "|", "^", ">>", "<<") {
// 			if stateT == 0 {
// 				opT[0] = nil
// 				opT[1] = v
// 				stateT = 2
// 			} else if stateT == 1 {
// 				opT[1] = v
// 				stateT = 2
// 			} else if stateT == 2 {
// 				opT[1] = v
// 				stateT = 2
// 			} else {
// 			}

// 		} else if strings.HasPrefix(v, "~") {
// 			if stateT == 0 {
// 				opT[0] = (*valuesA)[v]
// 				stateT = 1
// 			} else if stateT == 1 {
// 				opT[0] = (*valuesA)[v]
// 				stateT = 1
// 			} else if stateT == 2 {
// 				opT[2] = (*valuesA)[v]
// 				stateT = 3
// 			}

// 		} else {
// 			vT := p.ParseVar(v)
// 			vvT := p.GetVarValue(vT)

// 			if stateT == 0 {
// 				opT[0] = vvT
// 				stateT = 1
// 			} else if stateT == 1 {
// 				opT[0] = vvT
// 				stateT = 1
// 			} else if stateT == 2 {
// 				opT[2] = vvT
// 				stateT = 3
// 			}
// 		}

// 		if stateT == 3 {
// 			// opListT = append(opListT, opT)

// 			rvT := evalSingle(opT)

// 			if tk.IsError(rvT) {
// 				return rvT
// 			}

// 			opT[0] = rvT

// 			stateT = 1
// 		}

// 	}

// 	return opT[0]
// }

// func (p *XieVM) EvalExpression(strA string) (resultR interface{}) {
// 	strT := strA
// 	regexpT := regexp.MustCompile(`\([^\(]*?\)`)

// 	valuesT := make(map[string]interface{})

// 	var tmpv interface{}

// 	for {
// 		matchT := regexpT.FindStringIndex(strT)

// 		if matchT == nil {
// 			tmpv = p.EvalExpressionNoGroup(strT, &valuesT)

// 			if tk.IsError(tmpv) {
// 				tk.Pl("表达式计算失败：%v", tmpv)
// 			}

// 			break
// 		} else {
// 			tmpv = p.EvalExpressionNoGroup(strT[matchT[0]:matchT[1]][1:matchT[1]-matchT[0]-1], &valuesT)

// 			if tk.IsError(tmpv) {
// 				tk.Pl("表达式计算失败：%v", tmpv)
// 			}
// 		}

// 		keyT := "~" + tk.IntToStr(len(valuesT))

// 		valuesT[keyT] = tmpv

// 		strT = strT[0:matchT[0]] + " " + keyT + " " + strT[matchT[1]:len(strT)]
// 	}

// 	resultR = tmpv

// 	return

// 	// listT := strings.Split(strA, " ")

// 	// lenT := len(listT)

// 	// opListT := make([][]interface{}, lenT)

// 	// stateT := 0 // 0: initial, 1: first value ready, 2: operator ready, 3: second value ready

// 	// for i, v := range listT {
// 	// 	v = strings.TrimSpace(v)

// 	// 	if v == "" {
// 	// 		continue
// 	// 	}

// 	// 	opT := []interface{}{nil, nil, nil}

// 	// 	if tk.InStrings(v, "+", "-", "*", "/", "%", "!", "&&", "||", "==", "!=", ">", "<", ">=", "<=") {
// 	// 		if stateT == 0 {
// 	// 			opT[0] = nil
// 	// 			opT[1] = v
// 	// 			stateT = 2
// 	// 		}

// 	// 	} else {
// 	// 		vT := p.ParseVar(v)
// 	// 		vvT := p.GetVarValue(vT)

// 	// 		if stateT == 0 {
// 	// 			opT[0] = vvT
// 	// 			stateT = 1
// 	// 		}
// 	// 	}

// 	// 	if stateT == 3 {
// 	// 		opListT = append(opListT, opT)
// 	// 	}

// 	// }

// }

// var TimeFormat = "2006-01-02 15:04:05"
// var TimeFormatMS = "2006-01-02 15:04:05.000"
// var TimeFormatMSCompact = "20060102150405.000"
// var TimeFormatCompact = "20060102150405"
// var TimeFormatCompact2 = "2006/01/02 15:04:05"
// var TimeFormatDateCompact = "20060102"
// var ConstMapG map[string]interface{} = map[string]interface{}{
// 	"timeFormat":        "2006-01-02 15:04:05",
// 	"timeFormatCompact": "20060102150405",
// }

var GlobalsG *GlobalContext

func init() {
	// tk.Pl("init")

	// InstrCodeSet = make(map[int]string, 0)

	// for k, v := range InstrNameSet {
	// 	InstrCodeSet[v] = k
	// }

	GlobalsG = &GlobalContext{}

	GlobalsG.Vars = make(map[string]interface{}, 0)

	GlobalsG.Vars["backQuote"] = "`"

	GlobalsG.Vars["timeFormat"] = "2006-01-02 15:04:05"

	GlobalsG.Vars["timeFormatCompact"] = "20060102150405"

}

// w2d
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
		return tk.Undefined
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
