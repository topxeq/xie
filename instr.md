## 谢语言指令参考（Xielang Instruction Reference）

### --- 内部使用或调试相关 internal & debug related ---

**invalidInstr**: 表示无效的指令 invalid instruction, use internally to indicate invalid instr(s) while parsing commands

**version**: 获取当前谢语言版本号 get current Xielang version, return a string type value, if the result parameter not designated, it will be put to the global variable $tmp(and it's same for other instructions has result and not variable parameters)

**pass**: 没有任何作用的指令，一般用于占位 do nothing, useful for placeholder

**debug**: 输出调试信息 output the debug info

**debugInfo**: 获取调试信息 get the debug info

**varInfo**: 获取变量定义信息 get the information of the variables

**help**: not implemented

**onError**: set error handler

**dumpf**

**defer**: delay running an instruction, the instruction will be running by order(first in last out) when the function returns or the program exits, or error occurrs

**deferStack**: get defer stack info

**isUndef**: 判断变量是否未被声明（定义），第一个结果参数可省略，第二个参数是要判断的变量

**isDef**: 判断变量是否已被声明（定义），第一个结果参数可省略，第二个参数是要判断的变量

**isNil**: 判断变量是否是nil，第一个结果参数可省略，第二个参数是要判断的变量

**test**: for test purpose, check if 2 values are equal

**testByStartsWith**: for test purpose, check if first string starts with the 2nd

**testByReg**: for test purpose, check if first string matches the regex pattern defined by the 2nd string

**typeOf**: 获取变量或数值类型（字符串格式），省略所有参数表示获取看栈值（不弹栈）的类型

**layer**: 获取变量所处的层级（主函数层级为0，调用的第一个函数层级为1，再嵌套调用的为2，……）

### --- -- run code related ---

**loadCode**: 载入字符串格式的谢语言代码到当前虚拟机中（加在最后），出错则返回error对象说明原因

**loadGel**: 从网络载入谢语言函数（称为gel，凝胶，取其封装的意思），生成compiled对象，一般作为封装函数调用，建议用runCall或goRunCall调用（函数通过准全局变量inputL和outL进行出入参数的交互），出错则返回error对象说明原因；用法：loadGel http://example.com/gel/get1.xie -key=abc123，-key参数可以输入解密密钥，-file参数表示从本地文件读取（默认从远程读取也可以用file://协议从本地读取）

**compile**: compile a piece of code

**quickRun**: quick run a piece of code, in a new running context(but same VM, so the global values are accessible), use exit to exit the running context, no return value needed(only erturn error object or "undefined")

**runCode**: 运行一段谢语言代码，在新的虚拟机中执行，除结果参数（不可省略）外，第一个参数是字符串类型的代码或编译后代码（必选，后面参数都是可选），第二个参数为任意类型的传入虚拟机的参数（虚拟机内通过inputG全局变量来获取该参数），第三个参数可以是一个列表，键值对将依次传入新虚拟机作为全局变量，这两个参数（第二、三个）如果不需要可以传入$nilG，后面的参数可以是一个字符串数组类型的变量或者多个字符串类型的变量，虚拟机内通过argsG（字符串数组）来对其进行访问。返回值是虚拟机正常运行返回值，即$outG或exit加参数的返回值。

**runPiece**: run a piece of code, in current running context，运行一段谢语言代码，在当前的虚拟机和运行上下文中执行，结果参数可省略，第一个参数是字符串类型的代码或编译后代码。不需要返回值，仅当发生运行错误时返回error对象，否则返回undefined，

**extractRun**: extract a piece of instrs in a running-context to a new running-context

**extractCompiled**: extract a piece of instrs in a running-context to a compiled object

**len**: 获取字符串、列表、映射等的长度，参数全省略表示取弹栈值

**fatalf**: printf then exit the program(类似pl输出信息后退出程序运行)

**goto**: jump to the instruction line (often indicated by labels)

**jmp**

**wait**: 等待可等待的对象，例如waitGroup或chan，如果没有指定，则无限循环等待（中间会周期性休眠），用于等待用户按键退出或需要静止等待等场景；如果给出一个字符串，则输出字符串后等待输入（回车确认）后继续；如果是整数或浮点数则休眠相应的秒数后继续；

**exitL**: terminate the program(maybe quickDelegate), can with a return value(same as assign the semi-global value $outL)

**exitfL**: terminate the program(maybe quickDelegate), can with a return string value assembled like fatalf/sprintf(same as assign the semi-global value $outL), usage: exitfL "error is %v" err1

**exit**: terminate the program, can with a return value(same as assign the global value $outG)

### --- var related ---

**global**: define a global variable

**var**: define a local variable

**const**: 获取预定义常量

**nil**: make a variable nil

**ref**: -> 获取变量的引用（取地址）

**refVar**: -> 获取谢语言变量的引用（用于局部变量或全局变量）

**refNative**

**unref**: 对引用进行解引用

**unrefVar**: 对引用进行解引用（用于局部变量或全局变量）

**assignRef**: 根据引用进行赋值（将引用指向的变量赋值）

**assignRefVar**: 根据引用进行赋值（用于局部变量或全局变量）

### --- push/peek/pop stack related ---

**push**: push any value to stack

**peek**: peek the value on the top of the stack

**pop**: pop the value on the top of the stack

**getStackSize**

**clearStack**

**pushRun**: push any value to running context stack

**peekRun**: peek the value on the top of the running stack

**popRun**: pop the value on the top of the running stack

**getRunStackSize**

**clearRunStack**

### --- shared sync map(cross-VM) related ---

**getSharedMap**: 获取所有的列表项

**getSharedMapItem**: 获取全局映射变量，用法：getSharedMapItem $result key default，其中key是键名，default是可以省略的默认值（省略时如果没有值将返回undefined）

**getSharedMapSize**

**tryGetSharedMapItem**

**tryGetSharedMapSize**

**setSharedMapItem**: 设置全局映射变量，用法：setSharedMapItem key value

**trySetSharedMapItem**

**deleteSharedMapItem**

**tryDeleteSharedMapItem**

**clearSharedMap**

**tryClearSharedMap**

**lockSharedMap**

**tryLockSharedMap**

**unlockSharedMap**

**readLockSharedMap**

**tryReadLockSharedMap**

**readUnlockSharedMap**

**quickClearSharedMap**

**quickGetSharedMapItem**

**quickGetSharedMap**

**quickSetSharedMapItem**

**quickDeleteSharedMapItem**

**quickSizeSharedMap**

### --- assign related ---

**assign**: assignment, from local variable to global, assign value to local if not found

**=**

**assignGlobal**: 声明（如果未声明的话）并赋值一个全局变量

**assignFromGlobal**: 声明（如果未声明的话）并赋值一个局部变量，从全局变量获取值（如果需要的话）

**assignLocal**: 声明（如果未声明的话）并赋值一个局部变量

### --- if/else, switch related ---

**if**: usage: if $boolValue1 :labelForTrue :labelForElse

**ifNot**: usage: if @`$a1 == #i3` :+1 :+2

**ifEval**: 判断第一个参数（字符串类型）表示的表达式计算结果如果是true，则跳转到指定标号处

**ifEmpty**: 判断是否是空（值为undefined、nil、false、空字符串、小于等于0的整数或浮点数均会满足条件），是则跳转

**ifNilOrEmpty**

**ifEqual**: 判断是否相等，是则跳转

**ifNotEqual**: 判断是否不等，是则跳转

**ifErr**: if error or TXERROR string then ... else ...

**ifErrX**

**switch**: 用法：switch $variableOrValue $value1 :label1 $value2 :label2 ... :defaultLabel

**switchCond**: 用法：switch $condition1 :label1 $condition2 :label2 ... :defaultLabel

### --- compare related ---

**==**: 判断两个数值是否相等，无参数时，比较两个弹栈值，结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

**!=**: 判断两个数值是否不等，无参数时，比较两个弹栈值，结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

**<**: 判断两个数值是否是第一个数值小于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

**>**: 判断两个数值是否是第一个数值大于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

**<=**: 判断两个数值是否是第一个数值小于等于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

**>=**: 判断两个数值是否是第一个数值大于等于第二个数值，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

**cmp**: 比较两个数值，根据结果返回-1，0或1，分别表示小于、等于、大于，无参数时，比较两个弹栈值（注意弹栈值先弹出的为第二个待比较数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待比较数值

### --- operator related ---

**inc**: ++

**++**

**dec**: --

**--**

**add**: add 2 values

**+**

**sub**: 两个数值相减，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值

**-**

**mul**: 两个数值相乘，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值

*****

**div**: 两个数值相除，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值

**/**

**mod**: 两个数值做取模计算，无参数时，将两个弹栈值相加（注意弹栈值先弹出的为第二个数值），结果压栈；参数为1个时是结果参数，两个数值从堆栈获取；参数为2个时，表示两个数值，结果压栈；参数为3个时，第一个参数是结果参数，后两个为待计算数值

**%**

**adds**: 将多个参数进行相加

**!**: 取反操作符，对于布尔值取反，即true -> false，false -> true。对于其他数值，如果是未定义的变量（即Undefined），返回true，否则返回false

**not**: 逻辑非操作符，对于布尔值取反，即true -> false，false -> true，对于int、rune、byte等按位取反，即 0 -> 1， 1 -> 0

**&&**: 逻辑与操作符

**||**: 逻辑或操作符

**&**: bit and

**|**: bit or

**^**: bit xor

**&^**: bit and not

**?**: 三元操作符，用法示例：? $result $a $s1 "abc"，表示判断变量$a中的布尔值，如果为true，则结果为$s1，否则结果值为字符串abc，结果值将放入结果变量result中，如果省略结果参数，结果值将会存入$tmp

**ifThenElse**

**flexEval**: 计算一个表达式，支持普通语法，结果参数不可省略，之后第一个参数是表达式字符串，然后是0个或多个参数，在表达式中可以用v1、v2……来指代，表达式采用 github.com/antonmedv/expr 提供的表达式计算引擎，相关进一步文档也可以从这里获取，并可参照例子 flexEval.xie

**flexEvalMap**: 类似flexEval指令，区别是：flexEval后从第二个参数开始可以接受多个参数，并在表达式中以v1、v2这样来指代，而flexEvalMap则只允许有一个参数，需要是映射类型，这样可以直接用键名在表达式中引用这些变量

**eval**: 计算一个表达式

**quickEval**: quick eval an expression, use {} to contain an instruction(no nested {} allowed) that return result value in $tmp

**eval**

### --- func related ---

**call**: call a normal function, usage: call $result :func1 $arg1 $arg2...

// result value could not be omitted, use $drop if not neccessary
// all arguments/parameters will be put into the local variable "inputL" in the function
// and the function should return result in local variable "outL"
// use "ret $result" is a covenient way to set value of $outL and return from the function
**ret**: return from a normal function or a fast call function, while for normal function call, can with a paramter for set $outL

**sealCall**: new a VM to run a function, output/input through inputG & outG

// 封装函数，结果参数不可省略，第一个参数可以是代码或编译后、运行上下文，或者起始标号（此时第二个参数应为结束标号），后面参数都将传入行运函数中的$inputG中，出参通过$outG传出
**runCall**: 调用称为“行运函数”的代码块，在同一虚拟机、新建的运行上下文中调用函数，结果参数不可省略，第一个参数可以是代码或编译后、运行上下文，或者起始标号（此时第二个参数应为结束标号），后面参数都将传入行运函数中的$inputL中。行运函数会进行函数压栈（以便新的运行上下文中可以deferUpToRoot），入参通过$inputL访问，出参通过$outL传出

**goRunCall**: runCall in thread

**threadCall**: 并发调用函数，在新虚拟机中运行，函数体内无需返回outG、outL等参数，结果参数不可省略（但仅在调用函数启动线程时如遇错误返回error对象，后续因为是并发调用，返回值无意义），第一个参数如果是个运行上下文对象，后续参数都是传入参数（通过$inputG访问）；第一个参数如果是一个标号或整数，则还需要第二个标号或整数，分别表示并发函数的开始指令标号与结束指令标号；第一个参数也可以是字符串类型的源代码或编译后代码

**goCall**

**go**: 快速并发调用一个标号处的代码，该段代码应该使用exit命令来表示退出该线程

**fastCall**: 快速调用函数，第一个参数是跳转标号，后面的参数将被依次压栈，可以在函数体内弹栈使用；函数体内应使用fastRet指令返回 fast call function, no function stack used, no result value allowed, use stack or variables for input and output(parameters after the label parameter will be pushed to the stack in order), use fastRet or ret to return

**fastRet**: 从快速函数中返回 return from fast function, used with fastCall

### --- for/range related ---

**for**: for loop, usage: for @`$a < #i10` `++ $a` :cont1 :+1 , if the quick eval result is true(bool value), goto label :cont1, otherwise goto :+1(the next line/instr), the same as in C/C++ "for (; a < 10; a++) {...}"

**range**: usage: range 5 :+1 :breakRange1, range #J`[{"a":1,"b":2},{"c":3,"d":4}]` :range1 :+1

**getIter**: get i, v or k, v in range

### --- array/slice related ---

**newList**: 新建一个数组，后接任意个元素作为数组的初始项

**newArray**

**addItem**: 数组中添加项

**addArrayItem**: 数组中添加项

**addListItem**

**addStrItem**

**deleteItem**: 数组中删除项

**deleteArrayItem**

**deleteListItem**

**addItems**: 数组添加另一个数组的值

**addArrayItems**

**addListItems**

**getAnyItem**

**setAnyItem**

**getItem**

**getArrayItem**

**[]**

**setItem**: 修改数组中某一项的值

**setArrayItem**

**slice**: 对列表（数组）切片，如果没有指定结果参数，将改变原来的变量。用法示例：slice $list4 $list3 #i1 #i5，将list3进行切片，截取序号1（包含）至序号5（不包含）之间的项，形成一个新的列表，放入变量list4中。可以使用“-”来表示省略某个数值，例如：slice $list1 $argsT 2 -，表示截取从序号2到后面所有的项。

### --- control related ---

**continue**: continue the loop or range, PS "continue 2" means continue the upper loop in nested loop, "continue 1" means continue the upper of upper loop, default is 1 but could be omitted

**break**: break the loop or range, PS "break 2" means break the upper loop in nested loop

### --- map related ---

**setMapItem**: 设置映射项，用法：setMapItem $map1 Name "李白"

**deleteMapItem**: 删除映射项

**removeMapItem**

**getMapItem**: 获取指定序号的映射项，用法：getMapItem $result $map1 #i2，获取map1中的序号为2的项（即第3项），放入结果变量result中，如果有第4个参数则为默认值（没有找到映射项时使用的值），省略时将是undefined（可与全局内置变量$undefined比较）

**{}**

**getMapKeys**: 取所有的映射键名，可用于手工遍历等场景

**toOrderedMap**: 转换列表为有序列表

### --- object related ---

**new**: 新建一个数据或对象，第一个参数为结果放入的变量（不可省略），第二个为字符串格式的数据类型或对象名，后面是可选的0-n个参数，目前支持byte、int等，注意一般获得的结果是引用（或指针）

**method**: 对特定数据类型执行一定的方法，例如：method $result $str1 trimSet "ab"，将对一个字符串类型的变量str1去掉首尾的a和b字符，结果放入变量result中（注意，该结果参数不可省略，即使该方法没有返回数据，此时可以考虑用$drop）

**mt**

**member**: 获取特定数据类型的某个成员变量的值，例如：member $result $requestG "Method"，将获得http请求对象的Method属性值（GET、POST等），结果放入变量result中（注意，该结果参数不可省略，即使该方法没有返回数据，此时可以考虑用$drop）

**mb**

**mbSet**: 设置某个成员变量

**newObj**: 注：已废弃，合入new指令。新建一个对象，第一个参数为结果放入的变量，第二个为字符串格式的对象名，后面是可选的0-n个参数，目前支持string、any等

**setObjValue**: 设置对象本体值

**getObjValue**: 获取对象本体值

**getMember**: 获取对象成员值

**setMember**: 设置对象成员值

**callObj**: 调用对象方法

### --- string related ---

**backQuote**: 获取反引号字符串

**quote**: 将字符串进行转义（加上转义符，如“"”变成“\"”）

**unquote**: 将字符串进行解转义

**isEmpty**: 判断字符串是否为空

**是否空串**

**isEmptyTrim**: 判断字符串trim后是否为空

**strAdd**

**strSplit**: 按指定分割字符串分割字符串，结果参数不可省略，用法示例：strSplit $result $str1 "," 3，其中第3个参数可选（即可省略），表示结果列表最多的项数（例如为3时，将只按逗号分割成3个字符串的列表，后面的逗号将忽略；省略或为-1时将分割出全部）

**strSplitLines**: 按行分割字符串（根据\n来分割，\r将会被去除），用法示例：strSplitLines $result $str1

**strSplitByLen**: 按长度拆分一个字符串为数组，注意由于是rune，可能不是按字节长度，例： strSplitByLen $listT $strT 10，可以加第三个参数表示字节数不能超过多少，加第四个参数表示分隔符（遇上分隔符从分隔符后重新计算长度，也就是说分割长度可以超过指定的个数，一般用于有回车的情况）

**strJoin**: 连接一个字符串数组，中间以指定分隔符分隔

**strJoinLines**: 连接一个字符串数组，中间以指换行符（\n）分隔

**strReplace**: 字符串替换，用法示例：strReplace $result $str1 $find $replacement

**strReplaceIn**: 字符串替换，可同时替换多个子串，用法示例：strReplace $result $str1 $find1 $replacement1 $find2 $replacement2

**trim**: 字符串首尾去空白，非字符串将自动转换为字符串

**strTrim**

**去空白**

**trimSet**: 字符串首尾去指定字符，除结果参数外第二个参数（字符串类型）指定去掉那些字符

**trimSetLeft**: 字符串首去指定字符，除结果参数外第二个参数（字符串类型）指定去掉那些字符

**trimSetRight**: 字符串尾去指定字符，除结果参数外第二个参数（字符串类型）指定去掉那些字符

**trimPrefix**: 字符串首去指定字符串，除结果参数外第二个参数（字符串类型）指定去掉的子串，如果没有则返回原字符串

**trimSuffix**: 字符串尾去指定字符串，除结果参数外第二个参数（字符串类型）指定去掉的子串，如果没有则返回原字符串

**toUpper**: 字符串转为大写

**toLower**: 字符串转为小写

**strPad**: 字符串补零等填充操作，例如 strPad $result $strT #i5 -fill=0 -right=true，第一个参数是接收结果的字符串变量（不可省略），第二个是将要进行补零操作的字符串，第三个参数是要补齐到几位，默认填充字符串fill为字符串0，right（表示是否在右侧填充）为false（也可以直接写成-right），因此上例等同于strPad $result $strT #i5，如果fill字符串不止一个字符，最终补齐数量不会多于第二个参数指定的值，但有可能少

**strContains**: 判断字符串是否包含某子串

**strContainsIn**: 判断字符串是否包含任意一个子串，结果参数不可省略

**strCount**: 计算字符串中子串的出现次数

**strRepeat**: 重复字符串n次生成新的字符串

**strIn**: 判断字符串是否在一个字符串列表中出现，函数定义： 用法：strIn $result $originStr -it $sub1 "sub2"，第一个可变参数如果以“-”开头，将表示参数开关，-it表示忽略大小写，并且trim再比较（strA并不trim）

**inStrs**

**strStartsWith**: 判断字符串是否以某个子串开头

**strStartsWithIn**: 判断字符串是否以某个子串开头（可以指定多个子串，符合任意一个则返回true），结果参数不可省略

**strEndsWith**: 判断字符串是否以某个子串结束

**strEndsWithIn**: 判断字符串是否以某个子串结束（可以指定多个子串，符合任意一个则返回true），结果参数不可省略

### --- binary related ---

**bytesToData**

**dataToBytes**

**bytesToHex**: 以16进制形式输出字节数组

**bytesToHexX**: 以16进制形式输出字节数组，字节中间以空格分割

### --- thread related ---

**lock**: lock an object which is lockable

**unlock**: unlock an object which is unlockable

**lockN**: lock a global, internal, predefined lock in a lock pool/array, 0 <= N < 10

**unlockN**: unlock a global, internal, predefined lock, 0 <= N < 10

**tryLockN**: try lock a global, internal, predefined lock, 0 <= N < 10

**readLockN**: read lock a global, internal, predefined lock in a lock pool/array, 0 <= N < 10

**readUnlockN**: read unlock a global, internal, predefined lock, 0 <= N < 10

**tryReadLockN**: try read lock a global, internal, predefined lock, 0 <= N < 10

### --- time related ---

**now**: 获取当前时间

**nowStrCompact**: 获取简化的当前时间字符串，如20220501080930

**nowStr**: 获取当前时间字符串的正式表达

**nowStrFormal**: 获取当前时间字符串的正式表达

**nowTick**: 获取当前时间的Unix时间戳形式，为13位字符串形式（最后三位是毫秒）

**getNowTick**

**timestamp**

**nowTickInt**: 获取当前时间的Unix时间戳形式，为整数形式，单位纳秒

**getNowTickInt**

**nowUTC**

**timeSub**: 时间进行相减操作

**timeToLocal**

**timeToGlobal**

**getTimeInfo**

**timeAddDate**

**formatTime**: 将时间格式化为字符串，用法：formatTime $endT $endT "2006-01-02 15:04:05"

**timeToStr**

**toTime**: 参看10871， 将任意数值（可为字符串、时间戳字符串、时间等）转换为时间，字符串格式可以类似now、2006-01-02 15:04:05、20060102150405、2006-01-02 15:04:05.000，或10位或13位的Unix时间戳，可带可选参数-global、-defaultNow、-defaultErr、-defaultErrStr、-format=2006-01-02等，

**timeToTick**: 时间转时间戳，时间可为toTime指令中参数的格式，结果为13位字符串（毫秒级）

**timeToTickInt**: 时间转时间戳，时间可为toTime指令中参数的格式，结果为整数（纳秒级）

**tickToTime**: 时间戳转换为时间，如果参数是nil则返回当前时间，如果参数是整数，则按纳秒转换，如果是字符串，则可转换13位（毫秒）或10位（秒）的时间戳，此时如果转换失败则返回时间的零值（1970年...）

**tickIntToTime**: 时间戳转换为时间，输入为纳秒单位的整数

### --- math related ---

**abs**: 取绝对值

**ceil**: 向上取整

**floor**: 向下取整

**round**: 四舍五入

**max**: 取多个数的最大值，结果参数不可省略

**maxX**

**min**: 取多个数的最小值，结果参数不可省略

**minX**

**adjustFloat**: 将类似32.0000000004这种浮点计算误差值的位数去掉，结果参数不可省略，用法：adjustFloat @`#f0.65 - #f0.6` 10，第二个参数是整理到小数点后多少位，可以省略，默认是10

### --- command-line related ---

**getParam**: 获取指定序号的命令行参数，结果参数外第一个参数为list或strList类型，第二个为整数，第三个为默认值（字符串类型），例：getParam $result $argsG 2 ""

**获取参数**

**getSwitch**: 获取命令行参数中指定的开关参数，结果参数外第一个参数为list或strList类型，第二个为类似“-code=”的字符串，第三个为默认值（字符串类型），例：getSwitch $result $argsG "-code=" ""，将获取命令行中-code=abc的“abc”部分。

**ifSwitchExists**: 判断命令行参数中是否有指定的开关参数，结果参数外第一个参数为list或strList类型，第二个为类似“-verbose”的字符串，例：ifSwitchExists $result $argsG "-verbose"，根据命令行中是否含有-verbose返回布尔值true或false

**switchExists**

**ifSwitchNotExists**

**switchNotExists**

**parseCommandLine**: -> 分析命令行字符串，类似os.Args的获取过程

### --- print related ---

**pln**: same as println function in other languages

**plo**: print a value with its type

**plos**: 输出多个变量或数值的类型和值

**pr**: the same as print in other languages

**prf**: the same as printf in other languages

**pl**

**plNow**: 相当于pl，之前多输出一个时间

**plv**

**plvsr**: 输出多个变量或数值的值的内部表达形式，之间以换行间隔

**pv**: 输出变量值和名字

**plErr**: 输出一个error（表示错误的数据类型）信息

**plErrX**: 输出一个error（表示错误的数据类型）或TXERROR字符串信息

**plErrStr**: 输出一个TXERROR字符串（表示错误的字符串，以TXERROR:开头，后面一般是错误原因描述）信息

**spr**: 相当于其它语言的sprintf函数

### --- scan/input related ---

**scanf**: 相当于其它语言的scanf函数

**sscanf**: 相当于其它语言的sscanf函数

### --- convert related ---

**convert**: 转换数值类型，例如 convert $a int

**hex**: 16进制编码，对于数字高位在后

**hexb**: 16进制编码，对于数字高位在前

**unhex**: 16进制解码，结果是一个字节列表

**hexToBytes**

**toHex**: 任意数值16进制编码

**hexToByte**: 16进制编码转字节

**toBool**

**toByte**

**toRune**

**toInt**: 任意数值转整数，可带一个默认值（转换出错时返回该值），不带的话返回-1

**toFloat**

**toStr**

**toTime**: 将任意数值（可为字符串、时间戳字符串、时间等）转换为时间，字符串格式可以类似now、2006-01-02 15:04:05、20060102150405、2006-01-02 15:04:05.000，或10位或13位的Unix时间戳，可带可选参数-global、-defaultNow、-defaultErr、-defaultErrStr、-format=2006-01-02等，

**toAny**

### --- err string(TXERROR:) related ---

**isErrStr**: 判断是否是TXERROR字符串，用法：isErrStr $result $str1 $errMsg，第三个参数可选（结果参数不可省略），如有则当str1为TXERROR字符串时，会放入错误原因信息

**errStrf**: 生成TXERROR字符串，用法：errStrf $result "error: %v" $errMsg\

**getErrStr**: 获取TXERROR字符串中的错误原因信息（即TXERROR:后的内容）

**checkErrStr**: 判断是否是TXERROR字符串，是则退出程序运行

### --- error related / err string(with the prefix "TXERROR:" ) related ---

**isErr**: 判断是否是error对象，结果参数不可省略，除结果参数外第一个参数是需要确定是否是error的对象，第二个可选变量是如果是error时，包含的错误描述信息

**getErrMsg**: 获取error对象的错误信息

**isErrX**: 同时判断是否是error对象或TXERROR字符串，用法：isErrX $result $err1 $errMsg，第三个参数可选（结果参数不可省略），如有会放入错误原因信息

**checkErrX**: check if variable is error or err string, and terminate the program if true(检查后续变量或数值是否是error对象或TXERROR字符串，是则输出后中止)

**getErrStrX**: 获取error对象或TXERROR字符串中的错误原因信息（即TXERROR:后的内容）

**errf**: 生成错误对象，类似printf

### --- common related ---

**clear**: clear various object, and the object with Close method

**close**: 关闭文件等具有Close方法的对象

**resultf**: 生成一个TXResult表示通用结果的对象，JSON表达类似{"Status": "fail", "Value": "auth failed"}，Status一般只有success和fail两个取值，Value一般在fail时为失败原因，还可以有其他字段

**resultFromJson**: 生成一个TXResult表示通用结果的对象，JSON表达类似{"Status": "fail", "Value": "auth failed"}，Status一般只有success和fail两个取值，Value一般在fail时为失败原因，还可以有其他字段

**resultFromJSON**: 根据JSON生成TXResult对象，失败时返回error对象

### --- http request/response related ---

**writeResp**: 写一个HTTP请求的响应

**setRespHeader**: 设置一个HTTP请求的响应头，如setRespHeader $responseG "Content-Type" "text/json; charset=utf-8"

**writeRespHeader**: 写一个HTTP请求的响应头状态，如writeRespHeader $responseG #i200

**getReqHeader**: 获取一个HTTP请求的请求头信息

**genJsonResp**: 生成一个JSON格式的响应字符，用法：genJsonResp $result $requestG "success" "Test passed!"，结果格式类似{"Status":"fail", "Value": "network timeout"}，其中Status字段表示响应处理结果状态，一般只有success和fail两种，分别表示成功和失败，如果失败，Value字段中为失败原因，如果成功，Value中为空或需要返回的信息

**genResp**

**serveFile**

**newMux**: 新建一个HTTP请求处理路由对象，等同于 new mux

**setMuxHandler**: 设置HTTP请求路由处理函数，用法：setMuxHandler $muxT "/text1" $arg $text1，其中，text1是字符串形式的处理函数代码，arg是可以传入处理函数的一个参数，处理函数内可通过全局变量inputG来访问，另外还有全局变量requestG表示请求对象，responseG表示响应对象，reqNameG表示请求的子路径，paraMapG表示请求的URL（GET）参数或POST参数映射

**setMuxStaticDir**: 设置静态WEB服务的目录，用法示例：setMuxStaticDir $muxT "/static/" "./scripts" ，设置处理路由“/static/”后的URL为静态资源服务，第1个参数为newMux指令创建的路由处理器对象变量，第2个参数是路由路径，第3个参数是对应的本地文件路径，例如：访问 http://127.0.0.1:8080/static/basic.xie，而当前目录是c:\tmp，那么实际上将获得c:\tmp\scripts\basic.xie

**startHttpServer**: 启动http服务器，用法示例：startHttpServer $resultT ":80" $muxT ；可以后面加-go参数表示以线程方式启动，此时应注意主线程不要退出，否则服务器线程也会随之退出，可以用无限循环等方式保持运行

**startHttpsServer**: 启动https(SSL)服务器，用法示例：startHttpsServer $resultT ":443" $muxT /root/server.crt /root/server.key -go

### --- web related ---

**getWeb**: 发送一个HTTP网络请求，并获取响应结果（字符串格式），getWeb指令除了第一个参数必须是返回结果的变量，第二个参数是访问的URL，其他所有参数都是可选的，method可以是GET、POST等；encoding用于指定返回信息的编码形式，例如GB2312、GBK、UTF-8等；headers是一个JSON格式的字符串，表示需要加上的自定义的请求头内容键值对；参数中还可以有一个映射类型的变量或值，表示需要POST到服务器的参数，另外可加-bytes参数表示传回字节数组结果，用法示例：getWeb $resultT "http://127.0.0.1:80/xms/xmsApi" -method=POST -encoding=UTF-8 -timeout=15 -headers=`{"Content-Type": "application/json"}` $mapT

**getWebBytes**: 与getWeb相同，但获取结果为字节数组

**downloadFile**: 下载文件

**isHttps**: 判断一个网络请求是否是https的

**getResource**: 获取JQuery等常用的脚本或其他内置文本资源，一般用于服务器端提供内置的jquery等脚本嵌入，避免从互联网即时加载，第一个的参数是jquery.min.js等js文件的名称，内置资源中如果含有反引号，将被替换成~~~存储，使用getResource时将被自动替换回反引号

**getResourceRaw**: 与getResource作用类似，唯一区别是不将~~~替换回反引号

**getResourceList**: 获取可获取的资源名称列表

### --- html related ---

**htmlToText**: 将HTML转换为字符串，用法示例：htmlToText $result $str1 "flat"，第3个参数开始是可选参数，表示HTML转文本时的选项

### --- regex related ---

**regReplace**

**regReplaceAllStr**

**regFindAll**: 获取正则表达式的所有匹配，结果参数不可省略，匹配组号默认为0，即完整匹配，用法示例：regFindAll $result $str1 $regex1 $group

**regFindAllGroups**: 获取正则表达式的所有匹配，结果参数不可省略，结果是二维字符串数组，包含各个组，其中第0组是完整匹配，1开始是各个括号中的匹配组，用法示例：regFindAllGroups $result $str1 $regex1

**regFind**: 获取正则表达式的第一个匹配，用法示例：regFind $result $str1 $regex1 $group

**regFindFirst**

**regFindFirstGroup**: 获取正则表达式的第一个匹配，返回所有的匹配组为一个列表，其中第一项是完整匹配结果，第二项是第一个匹配组……，，用法示例：regFindFirstGroup $result $str1 $regex1

**regFindIndex**: 获取正则表达式的第一个匹配的位置，返回一个整数数组，任意值为-1表示没有找到匹配，用法示例：regFindIndex $result $str1 $regex1

**regMatch**: 判断字符串是否完全符合正则表达式，用法示例：regMatch $result "abcab" `a.*b`

**regContains**: 判断字符串中是否包含符合正则表达式的子串

**regContainsIn**: 判断字符串中是否包含符合任意一个正则表达式的子串

**regCount**: 计算字符串中包含符合某个正则表达式的子串个数，用法示例：regCount $result $str1 $regex1

**regSplit**: 用正则表达式分割字符串

**regQuote**: 将一个普通字符串中涉及正则表达式特殊字符进行转义替换以便用于正则表达式中

### --- system related ---

**sleep**: sleep for n seconds(float, 0.001 means 1 millisecond)

**getClipText**: 获取剪贴板文本

**setClipText**: 设置剪贴板文本

**getEnv**: 获取环境变量

**setEnv**: 设置环境变量

**removeEnv**: 删除环境变量

**systemCmd**: 执行一条系统命令，例如： systemCmd "cmd" "/k" "copy a.txt b.txt"

**openWithDefault**: 用系统默认的方式打开一个文件，例如： openWithDefault "a.jpg"

**getOSName**: 获取操作系统名称，如windows,linux,darwin等

**getOsName**

### --- file related ---

**loadText**: load text from file

**saveText**: 保存文本到指定文件

**loadBytes**: 从指定文件载入数据（字节列表）

**saveBytes**: 保存数据（字节列表）到指定文件

**loadBytesLimit**: 从指定文件载入数据（字节列表），不超过指定字节数

**appendText**: 追加文本到指定文件末尾

**writeStr**: 写入字符串，可以向文件、字节数组、字符串等写入，结果参数不可省略

**writeStrf**: 写入字符串，可以向文件、字节数组、字符串等写入，类似printf后的可变参数，结果参数不可省略

**readStr**: 读出字符串，可以从文件、字节数组、字符串等读取，读取所有能读取的

**readAllStr**: 读出字符串，可以从文件、字节数组、字符串等读取，读取所有能读取的

**createFile**: 新建文件，如果带-return参数，将在成功时返回FILE对象，失败时返回error对象，否则返回error对象，成功为nil，-overwrite有重复文件不会提示。如果需要指定文件标志位等，用openFile指令

**openFile**: 打开文件，如果带-readOnly参数，则为只读，-write参数可写，-create参数则无该文件时创建一个，-perm=0777可以指定文件权限标志位

**openFileForRead**: 打开一个文件，仅为读取内容使用

**closeFile**: 关闭文件

**readByte**: 从io.Reader或bufio.Reader读取一个字节

**readBytesN**: 从io.Reader或bufio.Reader读取多个字节，第二个参数指定所需读取的字节数

**writeByte**: 向io.Writer或bufio.Writer写入1个字节

**writeBytes**: 向io.Writer或bufio.Writer写入多个字节

**flush**: bufio.Writer等具有缓存的对象清除缓存

**cmpBinFile**: 逐个字节比较二进制文件，用法： cmpBinFile $result $file1 $file2 -identical -verbose，如果带有-identical参数，则只比较文件异同（遇上第一个不同的字节就返回布尔值false，全相同则返回布尔值true），不带-identical参数时，将返回一个比较结果对象

**fileExists**: 判断文件是否存在

**ifFileExists**

**isDir**: 判断是否是目录

**getFileSize**: 获取文件大小

**getFileInfo**: 获取文件信息，返回映射对象，参看genFileList命令

**removeFile**: 删除文件，用法：remove $rs $fileNameT -dry

**removeAllFile**: 删除所有文件（如果是目录名，将递归删除子目录下所有目录和文件），用法：removeAllFile $rs $fileNameT -dry

**renameFile**: 重命名文件

**copyFile**: 复制文件，用法 copyFile $result $fileName1 $fileName2，可带参数-force和-bufferSize=100000等

### --- path related ---

**genFileList**: 生成目录中的文件列表，即获取指定目录下的符合条件的所有文件，例：getFileList $result `d:\tmp` "-recursive" "-pattern=*" "-exclusive=*.txt" "-withDir" "-verbose"，另有 -compact 参数将只给出Abs、Size、IsDir三项, -dirOnly参数将只列出目录（不包含文件），列表项对象内容类似：map[Abs:D:\tmpx\test1.gox Ext:.gox IsDir:false Mode:-rw-rw-rw- Name:test1.gox Path:test1.gox Size:353339 Time:20210928091734]

**getFileList**

**joinPath**: join file paths

**getCurDir**: get current working directory

**setCurDir**: set current working directory

**getAppDir**: get the application directory(where execute-file exists)

**getConfigDir**: get application config directory

**extractFileName**: 从文件路径中获取文件名部分

**getFileBase**

**extractFileExt**: 从文件路径中获取文件扩展名（后缀）部分

**extractFileDir**: 从文件路径中获取文件目录（路径）部分

**extractPathRel**: 从文件路径中获取文件相对路径（根据指定的根路径）

**ensureMakeDirs**

### --- console related ---

**getInput**: 从命令行获取输入，第一个参数开始是提示字符串，可以类似printf加多个参数，用法：getInput $text1 "请输入%v个数字：" #i2

**getInputf**

**getPassword**: 从命令行获取密码输入（输入字符不显示），第一个参数是提示字符串

### --- json related ---

**toJson**: 将对象编码为JSON字符串

**toJSON**

**fromJson**: 将JSON字符串转换为对象

**fromJSON**

### --- xml related ---

**toXml**: 将对象编码为XML字符串

**toXML**

**getAllNodesTextFromXML**: 从XML文本中获取所有文本节点形成文本数组

### --- simple map related ---

**toSimpleMap**: 将对象编码为Simple Map字符串（每行一个键值对，以第一个等号分隔，例如：name=Zhang San）

**fromSimpleMap**: 将Simple Map字符串解码为映射对象

### --- random related ---

**randomize**: 初始化随机种子

**getRandomInt**: 获得一个随机整数，结果参数不可省略，此外带一个参数表示获取[0,参数1]之间的随机整数，带两个参数表示获取[参数1,参数2]之间的随机整数

**getRandomFloat**: 获得一个介于[0, 1)之间的随机浮点数

**genRandomStr**: 生成随机字符串，用法示例：genRandomStr $result -min=6 -max=8 -noUpper -noLower -noDigit -special -space -invalid，其中，除结果参数外所有参数均可选，-min用于设置最少生成字符个数，-max设置最多字符个数，-noUpper设置是否包含大写字母，-noLower设置是否包含小写字母，-noDigit设置是否包含数字，-special设置是否包含特殊字符，-space设置是否包含空格，-invalid设置是否包含一般意义上文件名中的非法字符，

**getRandomStr**

### --- encode/decode related ---

**md5**: 生成MD5编码

**simpleEncode**: 简单编码，主要为了文件名和网址名不含非法字符

**simpleDecode**: 简单编码的解码

**urlEncode**: URL编码（http://www.aaa.com -> http%3A%2F%2Fwww.aaa.com）

**urlDecode**: URL解码

**base64Encode**: Base64编码，输入参数是[]byte字节数组或字符串

**base64**

**toBase64**

**base64Decode**: Base64解码

**unbase64**

**fromBase64**

**htmlEncode**: HTML编码（&nbsp;等）

**htmlDecode**: HTML解码

**hexEncode**: 十六进制编码，仅针对字符串

**strToHex**

**hexDecode**: 十六进制解码，仅针对字符串

**hexToStr**

**toUtf8**: 转换字符串或字节列表为UTF-8编码，结果参数不可省略，第一个参数为要转换的源字符串或字节列表，第二个参数表示原始编码（默认为GBK）

**toUTF8**

### --- encrypt/decrypt related ---

**encryptText**: 用TXDEF方法加密字符串

**decryptText**: 用TXDEF方法解密字符串

**encryptData**: 用TXDEF方法加密数据（字节列表）

**decryptData**: 用TXDEF方法解密数据（字节列表）

### --- compress/uncompress related ---

**compress**: compress string or byte array to byte array

**uncompress**

**decompress**

**compressText**: compress string to string, may even be longer than original

**uncompressText**

**decompressText**

// network relate 网络相关
**getRandomPort**: 获取一个可用的socket端口（注意：获取后应尽快使用，否则仍有可能被占用）

**listen**: net.Listen

**accept**: net.Listener.Accept()

**startSocksServer**: 启动一个Socks5透传服务器，用法：startSocksServer $result -ip=0.0.0.0 -port=8080 -password=acb123 -verbose，参数都是可选

**startSocksClient**: 启动一个Socks5透传客户端，用法：startSocksClient $result -remoteIp=0.0.0.0 -remotePort=8080 -localIp=0.0.0.0 -localPort=8081 -password=acb123 -verbose，参数都是可选

### --- database related ---

**dbConnect**: 连接数据库，用法示例：dbConnect $db "sqlite3" `c:\tmpx\test.db`，或dbConnect $db "godror" `user/pass@129.0.9.11:1521/testdb`，结果参数外第一个参数为数据库驱动类型，目前支持sqlite3、mysql、mssql、godror（即oracle）等，第二个参数为连接字串

**连接数据库**

**dbClose**: 关闭数据库连接

**关闭数据库**

**dbQuery**: 在指定数据库连接上执行一个查询的SQL语句（一般是select等），返回数组，每行是映射（字段名：字段值），用法示例：dbQuery $rs $db $sql $arg1 $arg2 ...

**查询数据库**

**dbQueryMap**: 在指定数据库连接上执行一个查询的SQL语句（一般是select等），返回一个映射，以指定的数据库记录字段为键名，对应记录为键值，用法示例：dbQueryMap $rs $db $sql $key $arg1 $arg2 ...

**查询数据库映射**

**dbQueryRecs**: 在指定数据库连接上执行一个查询的SQL语句（一般是select等），返回二维数组（第一行为字段名），用法示例：dbQueryRecs $rs $db $sql $arg1 $arg2 ...

**查询数据库记录**

**dbQueryCount**: 在指定数据库连接上执行一个查询的SQL语句，返回一个整数，一般用于查询记录条数或者某个整数字段的值等场景，用法示例：dbQueryCount $rs $db `select count(*) from TABLE1 where FIELD1=:v1 and FIELD2=:v2` $arg1 $arg2 ...

**dbQueryFloat**: 在指定数据库连接上执行一个查询的SQL语句，返回一个浮点数，用法示例：dbQueryFloat $rs $db `select PRICE from TABLE1 where FIELD1=:v1 and FIELD2=:v2` $arg1 $arg2 ...

**dbQueryString**: 在指定数据库连接上执行一个查询的SQL语句，返回一个字符串，用法示例：dbQueryString $rs $db `select NAME from TABLE1 where FIELD1=:v1 and FIELD2=:v2` $arg1 $arg2 ...

**dbQueryMapArray**: 在指定数据库连接上执行一个查询的SQL语句（一般是select等），返回一个映射，以指定的数据库记录字段为键名，对应记录为键值的数组，用法示例：dbQueryMapArray $rs $db $sql $key $arg1 $arg2 ...

**dbQueryOrdered**: 同dbQuery，但返回一个有序列表（orderedMap）

**dbExec**: 在指定数据库连接上执行一个有操作的SQL语句（一般是insert、update、delete等），用法示例：dbExec $rs $db $sql $arg1 $arg2 ...

**执行数据库**

### --- markdown related ---

**renderMarkdown**: 将Markdown格式字符串渲染为HTML

### --- image related ---

**newImage**

**pngEncode**: 将图像保存为PNG文件或其他可写载体（如字符串），用法：pngEncode $errT $fileT $imgT

**jpgEncode**: 将图像保存为JPG文件或其他可写载体（如字符串），用法：jpgEncode $errT $fileT $imgT -quality=90

**jpegEncode**

### --- screen related ---

**getActiveDisplayCount**: 获取活跃屏幕数量

**getScreenResolution**: 获取指定屏幕的分辨率，用法：getScreenResolution $rectT -index=0 -format=rect，其中后面的参数均可选，index指定要获取的活跃屏幕号，主屏幕是0，format可以是rect、json或为空，参看例子代码getScreenInfo.xie

**captureDisplay**: 屏幕截图，用法 captureDisplay $imgT 0，截取0号活跃屏幕（主屏幕）的全屏截图

**captureScreen**: 屏幕区域截图，用法 captureScreenRect $imgT 100 100 640 480，截取主屏幕的坐标(100,100)为左上角，宽640，高480的区域截图，后面几个参数均可省略，默认截全屏

**captureScreenRect**

### --- token related ---

**genToken**: 生成令牌，用法：genToken $result $appCode $userID $userRole -secret=abc，其中可选开关secret是加密秘钥，可省略

**checkToken**: 检查令牌，用法：checkToken $result XXXXX -secret=abc -expire=2，其中expire是设置的超时秒数（默认为1440），如果成功，返回类似“appCode|userID|userRole|”的字符串；失败返回TXERROR字符串

### --- line editor related ---

**leClear**: 清空行文本编辑器缓冲区

**leLoadStr**: 行文本编辑器缓冲区载入指定字符串内容，例：leLoadStr $textT "abc\nbbb\n结束"

**leSetAll**: 等同于leLoadStr

**leSaveStr**: 取出行文本编辑器缓冲区中内容，例：leSaveStr $result $s

**leGetAll**: 等同于leSaveStr

**leLoad**: 从文件中载入文本到行文本编辑器缓冲区中，例：leLoad $result `c:\test.txt`

**leLoadFile**: 等同于leLoad

**leLoadClip**: 从剪贴板中载入文本到行文本编辑器缓冲区中，例：leLoadClip $result

**leLoadSSH**: 从SSH连接获取文本文件内容，用法：leLoadSSH 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -path=远端文件路径，结果变量不可省略，其他参数省略时将从之前获取内容的SSH连接中获取

**leLoadUrl**: 从网址URL载入文本到行文本编辑器缓冲区中，例：leLoadUrl $result `http://example.com/abc.txt`

**leSave**: 将行文本编辑器缓冲区中内容保存到文件中，例：leSave $result `c:\test.txt`

**leSaveFile**: 等同于leSave

**leSaveClip**: 将行文本编辑器缓冲区中内容保存到剪贴板中，例：leSaveClip $result

**leSaveSSH**: 将编辑缓冲区内容保存到SSH连接中，如果不带参数，将从之前获取内容的SSH连接中获取，用法：leSaveSSH 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -path=远端文件路径。结果参数不可省略

**leInsert**: 行文本编辑器缓冲区中的指定位置前插入指定内容，例：leInsert $result 3 "abc"

**leInsertLine**: 等同于leInsert

**leAppend**: 行文本编辑器缓冲区中的最后追加指定内容，例：leAppendLine $result "abc"

**leAppendLine**: 等同于leAppend

**leSet**: 设定行文本编辑器缓冲区中的指定行为指定内容，例：leSet  $result 3 "abc"

**leSetLine**: 设定行文本编辑器缓冲区中的指定行为指定内容，例：leSetLine $result 3 "abc"

**leSetLines**: 设定行文本编辑器缓冲区中指定范围的多行为指定内容，例：leSetLines $result 3 5 "abc\nbbb"

**leRemove**: 删除行文本编辑器缓冲区中的指定行，例：leRemove $result 3

**leRemoveLine**: 等同于leRemove

**leRemoveLines**: 删除行文本编辑器缓冲区中指定范围的多行，例：leRemoveLines $result 1 3

**leViewAll**: 查看行文本编辑器缓冲区中的所有内容，例：leViewAll $textT

**leView**: 查看行文本编辑器缓冲区中的指定行，例：leView $lineText 18

**leViewLine**

**leSort**: 将行文本编辑器缓冲区中的行进行排序，唯一参数（可省略，默认为false）表示是否降序排序，例：leSort $result true

**leEnc**: 将行文本编辑器缓冲区中的文本转换为UTF-8编码，如果不指定原始编码则默认为GB18030编码，用法：leEnc $result gbk

**leLineEnd**: 读取或设置行文本编辑器缓冲区中行末字符（一般是\n或\r\n），不带参数是获取，带参数是设置

**leSilent**: 读取或设置行文本编辑器的静默模式（布尔值），不带参数是获取，带参数是设置

**leFind**: 在编辑缓冲区查找包含某字符串（可以是正则表达式）的行

**leReplace**: 在编辑缓冲区查找包含某字符串（可以是正则表达式）的行并替换相关内容

**leSSHInfo**: 获取当前行文本编辑器使用的SSH连接的信息

**leRun**: 将当前行文本编辑器中保存的文本作为谢语言代码执行，结果参数不可省略，如有第二个参数则为要传入的inputG，第三个参数开始为传入的argsG

### --- server related ---

**getMimeType**: 根据文件名获取MIME类型，文件名可以包含路径

**getMIMEType**

### --- zip related ---

**archiveFilesToZip**: 添加多个文件到一个新建的zip文件，第一个参数为zip文件名，后缀必须是.zip，可选参数-overwrite（是否覆盖已有文件），-makeDirs（是否根据需要新建目录），其他参数看做是需要添加的文件或目录，目录将递归加入zip文件，如果参数为一个列表，将看作一个文件名列表，其中的文件都将加入

**extractFilesFromZip**: 添加文件到zip文件

**compressData**: 压缩数据，用法：compressData $result $data -method=gzip，压缩方法由-method参数指定，默认为gzip，还支持lzw

**decompressData**: 解压缩数据

### --- web GUI related ---

**initWebGUIW**: 初始化Web图形界面编程环境（Windows下IE11版本），如果没有外嵌式浏览器xiewbr，则将其下载到xie语言目录下

**initWebGuiW**

**updateWebGuiW**: 强制刷新Web图形界面编程环境（Windows下IE11版本），会将最新的外嵌式浏览器xiewbr下载到xie语言目录下

**initWebGUIC**: 初始化Web图形界面编程环境（Windows下CEF版本），如果没有外嵌式浏览器xiecbr及相关库文件，则将其下载到xie语言目录下

**initWebGuiC**

### --- ssh/sftp/ftp related ---

**sshConnect**: 打开一个SSH连接，用法：sshConnect 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码

**sshOpen**

**sshClose**: 关闭一个SSH连接

**sshUpload**: 通过ssh上传一个文件，用法：sshUpload 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -path=文件路径 -remotePath=远端文件路径，可以加-force参数表示覆盖已有文件

**sshUploadBytes**: 通过ssh上传一个二进制内容（字节数组）到文件，用法：sshUpload 结果变量 内容变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -remotePath=远端文件路径，可以加-force参数表示覆盖已有文件；内容变量也可以是一个字符串，将自动转换为字节数组

**sshDownload**: 通过ssh下载一个文件，用法：sshDownload 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -path=本地文件路径 -remotePath=远端文件路径，可以加-force参数表示覆盖已有文件

**sshDownloadBytes**: 通过ssh下载一个文件，结果为字节数组（或error对象），用法：sshDownloadBytes 结果变量 -host=服务器名 -port=服务器端口 -user=用户名 -password=密码 -remotePath=远端文件路径，可以加-force参数表示覆盖已有文件

### --- excel related ---

**excelNew**: 新建一个excel文件，用法：excelNew $excelFileT

**excelOpen**: 打开一个excel文件，用法：excelOpen $excelFileT `d:\tmp\excel1.xlsx`

**excelClose**: 关闭一个excel文件

**excelSaveAs**: 保存一个excel文件，用法：excelSave $result $excelFileT `d:\tmp\excel1.xlsx`

**excelWrite**: 将excel文件内容写入到可写入源（io.Writer），例如文件、标准输出、网页输出http的response excelWrite $result $excelFileT $writer

**excelReadSheet**: 读取已打开的excel文件某一个sheet的内容，返回格式是二维数组，用法：excelReadSheet $result $excelFileT sheet1，最后一个参数可以是字符串类型表示按sheet名称读取，或者是一个整数表示按序号读取

**excelReadCell**: 读取指定单元格的内容，返回字符串或错误信息，用法：excelReadCell $result $excelFileT "sheet1" "A1"

**excelWriteCell**: 将内容写入到指定单元格，用法：excelWriteCell $result $excelFileT "sheet1" "A1" "abc123"

**excelSetCell**

**excelGetSheetList**: 获取sheet名字列表，结果是字符串数组

### --- mail related ---

**mailNewSender**: 新建一个邮件发送对象，与 new $result "mailSender" 指令效果类似

### --- pinyin related ---

**toPinYin**: 字符串转换为拼音，结果参数不可省略，用法：toPinyin $result "我们都是nice的。"，结果是所有汉字转为的拼音和所有无法转为拼音的字符的原字符；可以加-sep=-表示将各个拼音和字符间以指定分隔符分隔，加-pinYinOnly开关参数表示只包含能够转换为拼音的字符，加-ascOnly表示只包含ASCII字符，加-first表示拼音只取首字母，-tone表示加音调，-digitTone表示音调以数字表示，-digitTone2表示音调以数字表示且加在韵母（元音）后，加-raw表示结果为二维字符串数组，加参数用法类似：toPinyin $pln "我们都是nice的。" -digitTone -sep=-

**toPinyin**

### --- misc related ---

**awsSign**

### --- GUI related ---

**guiInit**: 初始化GUI环境

**alert**: 类似JavaScript中的alert，弹出对话框，显示一个字符串或任意数字、对象的字符串表达

**guiAlert**

**msgBox**: 类似Delphi、VB中的msgBox，弹出带标题的对话框，显示一个字符串，第一个参数是标题，第二个是字符串

**showInfo**

**guiShowInfo**

**showError**: 弹框显示错误信息

**guiShowError**

**getConfirm**: 显示信息，获取用户的确认

**guiGetConfirm**

**guiNewWindow**

**guiMethod**: 调用GUI生成的对象的方法

**guiMt**

