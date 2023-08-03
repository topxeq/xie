谢语言是一门免费、开源、跨平台、跨语言、语法接近汇编语言与SHELL脚本、全栈、易嵌入、快速的解释性计算机编程语言。

Xielang is a free, open-source, cross-platform, cross-language, ASM/SHELL-like, embeddable, full-stack, fast scripting language.

- [介绍（Introduction）](#介绍introduction)
- [语言特点（Features）](#语言特点features)
- [语言设计构思（Language design conception）](#语言设计构思language-design-conception)
- [安装教程（Installation）](#安装教程installation)
- [代码编辑器（Code Editor）](#代码编辑器code-editor)
- [运行和查看例子代码（Run and view sample code）](#运行和查看例子代码run-and-view-sample-code)
- [快速入门及语言要点（Quick Tour and Language Essentials）](#快速入门及语言要点quick-tour-and-language-essentials)
  - [- **基本语法**（Basic grammar）](#--基本语法basic-grammar)
  - [- **代码注释**（Comments）](#--代码注释comments)
  - [- **变量声明（定义）**（Variable declaration/definition）](#--变量声明定义variable-declarationdefinition)
  - [- **数值与类型**（Value and types）](#--数值与类型value-and-types)
  - [- **给变量赋值**（Assignment）](#--给变量赋值assignment)
  - [- **指定赋值的类型**（Specify the type of assignment）](#--指定赋值的类型specify-the-type-of-assignment)
  - [- **字符串赋值**（String assignment）](#--字符串赋值string-assignment)
  - [- **各种赋值示例**（Examples of various assignments）](#--各种赋值示例examples-of-various-assignments)
  - [- **用var指令的时候赋值**（Assign value when using "var" instruction）](#--用var指令的时候赋值assign-value-when-using-var-instruction)
  - [- **堆栈**（Stack）](#--堆栈stack)
  - [- **基本的堆栈操作**（Basic stack operations）](#--基本的堆栈操作basic-stack-operations)
  - [- **与堆栈有关的特殊变量**（Special variables related to stack）](#--与堆栈有关的特殊变量special-variables-related-to-stack)
  - [- **常见运算**（Common operations）](#--常见运算common-operations)
  - [- **数值类型转换**（Data type conversion）](#--数值类型转换data-type-conversion)
  - [- **字符串的连接操作**（String connection operation）](#--字符串的连接操作string-connection-operation)
  - [- **指令的结果参数**（Result parameter of instruction）](#--指令的结果参数result-parameter-of-instruction)
  - [- **pl指令**（ the "pl" instr）](#--pl指令-the-pl-instr)
  - [- **内置全局变量**（Predefined/built-in global variables）](#--内置全局变量predefinedbuilt-in-global-variables)
  - [- **标号**（Labels）](#--标号labels)
  - [- **代码缩进**（Code Indent）](#--代码缩进code-indent)
  - [- **复杂表达式分解**（Complex expression decomposition）](#--复杂表达式分解complex-expression-decomposition)
  - [- **复杂表达式运算**（Complex expression operation）](#--复杂表达式运算complex-expression-operation)
  - [- **复杂表达式做参数**](#--复杂表达式做参数)
  - [- **表达式的另一个例子**（Another example of an expression）](#--表达式的另一个例子another-example-of-an-expression)
  - [- **灵活表达式**（Flex expression）](#--灵活表达式flex-expression)
  - [- **灵活表达式做参数**（Flexible expression as parameter）](#--灵活表达式做参数flexible-expression-as-parameter)
  - [- **goto语句**（The goto instr）](#--goto语句the-goto-instr)
  - [- **一般循环结构**（Loop cycle structure）](#--一般循环结构loop-cycle-structure)
  - [- **条件分支**（Conditional branch）](#--条件分支conditional-branch)
  - [- **else分支**（Else branch）](#--else分支else-branch)
  - [- **虚拟标号/伪标号跳转**（Virtual label/pseudolabel jump）](#--虚拟标号伪标号跳转virtual-labelpseudolabel-jump)
  - [- **for循环**（The for loop）](#--for循环the-for-loop)
  - [- **利用for指令进行for循环**（Use the for instruction to carry out a for loop）](#--利用for指令进行for循环use-the-for-instruction-to-carry-out-a-for-loop)
  - [- **用range指令进行简单数据的遍历**（Iterating data using the range instruction）](#--用range指令进行简单数据的遍历iterating-data-using-the-range-instruction)
  - [- **range嵌套**（range in range）](#--range嵌套range-in-range)
  - [- **更多range数字的例子**（range numbers）](#--更多range数字的例子range-numbers)
  - [- **switch分支**（switch branches）](#--switch分支switch-branches)
  - [- **switchCond分支**（switchCond branches）](#--switchcond分支switchcond-branches)
  - [- **函数调用**（Function call）](#--函数调用function-call)
  - [- **函数调用时传递参数**（passing/retieving parameters in function call）](#--函数调用时传递参数passingretieving-parameters-in-function-call)
  - [- **全局变量和局部变量**（Global and local variables）](#--全局变量和局部变量global-and-local-variables)
  - [- **快速函数**（Fast functions）](#--快速函数fast-functions)
  - [- **寄存器**（Registers）](#--寄存器registers)
  - [- **谢语言的基础设施**（Infrastructure provided by Xielang）](#--谢语言的基础设施infrastructure-provided-by-xielang)
  - [- **用runCall指令在不同运行上下文中执行代码**（Executing code in different running contexts using the runCall instruction）](#--用runcall指令在不同运行上下文中执行代码executing-code-in-different-running-contexts-using-the-runcall-instruction)
  - [- **取变量引用及取引用对应变量的实际值**（Reference and Dereference）](#--取变量引用及取引用对应变量的实际值reference-and-dereference)
  - [- **取谢语言变量的引用及其解引用**（Reference variable in Xielang and Dereference it）](#--取谢语言变量的引用及其解引用reference-variable-in-xielang-and-dereference-it)
  - [- **复杂数据类型-列表**（Complex Data Types - List）](#--复杂数据类型-列表complex-data-types---list)
  - [- **复杂数据类型-映射**（Complex Data Types - Map）](#--复杂数据类型-映射complex-data-types---map)
  - [- **取列表项、映射项和列表切片的快捷写法**（Shortcut for taking list items, mapping items, and list slices）](#--取列表项映射项和列表切片的快捷写法shortcut-for-taking-list-items-mapping-items-and-list-slices)
  - [- **嵌套的复杂数据结构及JSON编码**（Nested complex data structures and JSON encoding）](#--嵌套的复杂数据结构及json编码nested-complex-data-structures-and-json-encoding)
  - [- **JSON解码**（JSON decoding）](#--json解码json-decoding)
  - [- **加载外部模块**（Loading external module）](#--加载外部模块loading-external-module)
  - [- **封装函数调用**（Sealed Function Call）](#--封装函数调用sealed-function-call)
  - [- **并发函数**（Concurrent function call）](#--并发函数concurrent-function-call)
  - [- **用线程锁处理并发共享冲突**（Using Thread Locks to Handle Concurrent Sharing Conflicts）](#--用线程锁处理并发共享冲突using-thread-locks-to-handle-concurrent-sharing-conflicts)
  - [- **对象机制**(Object Model)](#--对象机制object-model)
  - [- **快速/宿主对象机制**(Fast/Host Object Mechanism)](#--快速宿主对象机制fasthost-object-mechanism)
  - [- **时间处理**(Time processing)](#--时间处理time-processing)
  - [- **错误处理**(Error Handling)](#--错误处理error-handling)
  - [- **延迟执行指令 defer**](#--延迟执行指令-defer)
  - [- **关系数据库访问**（Relational Database Access）](#--关系数据库访问relational-database-access)
  - [- **微服务/应用服务器**（Microservices/Application Server）](#--微服务应用服务器microservicesapplication-server)
  - [- **网络（HTTP）客户端**（Network(HTTP) Client）](#--网络http客户端networkhttp-client)
  - [- **手动编写Api服务器**（Manually writing Api servers）](#--手动编写api服务器manually-writing-api-servers)
  - [- **静态WEB服务器**（Static WEB server）](#--静态web服务器static-web-server)
  - [- **动态网页服务器**（Dynamic Web Server）](#--动态网页服务器dynamic-web-server)
  - [- **博客系统**（Implementing a tiny blog system）](#--博客系统implementing-a-tiny-blog-system)
  - [- **嵌套运行谢语言代码**（Nested Run Xielang Code）](#--嵌套运行谢语言代码nested-run-xielang-code)
- [**谢语言做系统服务**（Write system services by Xielang）](#谢语言做系统服务write-system-services-by-xielang)
- [**图形界面（GUI）编程**（GUI Programming）](#图形界面gui编程gui-programming)
- [**谢语言GUI编程的基础（WebView2）** （Fundamentals of GUI Programming in Xielang (WebView2)）](#谢语言gui编程的基础webview2-fundamentals-of-gui-programming-in-xielang-webview2)
  - [- **基本界面**（Basic GUI）](#--基本界面basic-gui)
  - [- **直接嵌入网页脚本**（Directly embed JavaScript file in WEB pages）](#--直接嵌入网页脚本directly-embed-javascript-file-in-web-pages)
  - [- **启动后台服务与前台配合**（Start the backend service and cooperate with the front end）](#--启动后台服务与前台配合start-the-backend-service-and-cooperate-with-the-front-end)
  - [- **简单的图形计算器**（A Simple GUI Calculator）](#--简单的图形计算器a-simple-gui-calculator)
  - [- **Windows编译不带命令行窗口的谢语言主程序**（Compiling Xielang main program without command line window for Windows）](#--windows编译不带命令行窗口的谢语言主程序compiling-xielang-main-program-without-command-line-window-for-windows)
  - [- **制作一个登录框**（Create a login box）](#--制作一个登录框create-a-login-box)
- [编译运行谢语言代码（Compile and run Xielang code）](#编译运行谢语言代码compile-and-run-xielang-code)
- [内置指令/命令/函数参考（Built-in instruction/command/function reference）](#内置指令命令函数参考built-in-instructioncommandfunction-reference)
- [内置对象参考（Built-in object reference）](#内置对象参考built-in-object-reference)
- [杂项说明（Miscellaneous description）](#杂项说明miscellaneous-description)
  - [- **指令的参数**（Parameter of instruction）](#--指令的参数parameter-of-instruction)
  - [- **行末注释**（Comment at the end of the line）](#--行末注释comment-at-the-end-of-the-line)
  - [- **自动执行**（Auto-run scripts）](#--自动执行auto-run-scripts)
  - [- **从剪贴板执行代码**（Run Xielang code from clipboard）](#--从剪贴板执行代码run-xielang-code-from-clipboard)
  - [- **指令参数中引号的位置**（The position of quotation marks in instruction parameters）](#--指令参数中引号的位置the-position-of-quotation-marks-in-instruction-parameters)
  - [- **fastCall指令调用的快速函数代码中使用+1等虚拟标号**（The "+1" virtual label used in the fast function code called by the fastCall instruction）](#--fastcall指令调用的快速函数代码中使用1等虚拟标号the-1-virtual-label-used-in-the-fast-function-code-called-by-the-fastcall-instruction)
- [性能方面的考虑（Performance considerations）](#性能方面的考虑performance-considerations)
- [嵌入式使用谢语言（以虚拟机的方式在其他语言中调用）（Embedded Xielang in other languages）](#嵌入式使用谢语言以虚拟机的方式在其他语言中调用embedded-xielang-in-other-languages)
- [扩展谢语言（Extended Xielang）](#扩展谢语言extended-xielang)
- [编译谢语言（Compile Xielang）](#编译谢语言compile-xielang)
- [代码示例（Code examples）](#代码示例code-examples)
- [参与贡献者（Contributors）](#参与贡献者contributors)



&nbsp;

#### 介绍（Introduction）

&nbsp;

谢语言（英文名称为：Xielang，官网 [xie.topget.org](http://xie.topget.org/)）是一门开源、免费的解释型编程语言（也称作脚本语言），最大的特色包括：跨平台；跨语言（目前支持Go语言、JavaScript语言等，即将支持Java语言）可嵌入（即可在这些语言中调用）；结合了汇编语言和高级语言的优点；支持全中文编程（包括提示信息），语法简单易懂；单文件无依赖；可编译成单独可执行文件发布等。

Xielang(official website [xie.topget.org](http://xie.topget.org/)) is an open source and free interpretative programming language (also known as script language). Main features include: cross-platform, cross language (currently supports Go language, JavaScript language, etc., and will soon support Java language), embeddable (in other languages), combines the advantages of assembly language and high-level language, and the syntax is simple, ASM/shell-script like and easy to rewrite in any other languages; minimum dependency, a single executable main program, code can be compiled into a separate executable file and distributed.

谢语言支持各种基本的语法元素和结构，包括变量、条件分支、循环、函数、递归函数调用、多线程等，支持作为嵌入型语言在不同语言中调用，也支持独立运行（单文件的可执行程序），还支持作为后台微服务运行。同时，谢语言也提供一个命令行交互式编程环境，可用于一般的测试。

Xielang supports various basic syntax elements and structures, including variables, conditional branches, loops, functions, recursive function calls, multi-threading, etc. It supports calls in other languages as embedded languages, minimum runtime dependency (single file executable), and supports system-service mode and micro-service mode. At the same time, Xielang also provides a command-line interactive programming environment, which can be used for general testing.

谢语言的Go语言版本，单文件即可执行，包含了脚本执行功能（无需安装其他依赖环境）、交互式命令行环境和微服务器模式，支持图形界面编程（可采用多种方式，无需或仅需附加一个动态链接库文件）。

The Go language version of Xielang, which can be executed in a single file, includes script execution function (no need to install other dependent environments), interactive command-line environment and micro-service mode, and supports graphical interface programming (GUI, multiple methods can be adopted, without or only need to attach a dynamic link library file).

谢语言的JavaScript版本，使用时仅需在网页中引用两个JavaScript文件，即可使用谢语言的功能，并且可以与JavaScript良好互通，充分发挥JavaScript中既有功能以及丰富的第三方库的优势。

The JavaScript version of Xielang can use all the functions of Xielang only by referencing two JavaScript files in the web page, and it can communicate well with JavaScript, benefit of the existing functions in JavaScript and the advantages of rich third-party libraries.

&nbsp;

#### 语言特点（Features）

&nbsp;

谢语言特点比较鲜明：

Xielang has distinctive features:

- 语法形式追求极简，类似命令行（Shell脚本）和汇编语言，基本结构是一行一条指令，以追求解析的速度，避免繁琐的语法、语义分析；

- The syntax form pursues minimalism, similar to command line (shell script) and assembly language. The basic structure is one instruction line by line, in order to pursue the speed of parsing and avoid tedious syntax and semantic analysis;
&nbsp;<br>

- 语法接近于汇编语言，包括一些语言的基础设施（例如堆栈、寄存器等）和语法结构（条件分支与无条件跳转指令等）；
- Syntax is close to assembly language, including some language infrastructure (such as stack, register, etc.) and syntax structure (conditional branch and unconditional jump instruction, etc.);
&nbsp;<br>

- 提供很多封装好的、功能丰富的内置指令，因而又使得语言接近于高级语言；
- It provides many encapsulated and functional built-in instructions, thus making the language close to high-level language;
&nbsp;<br>

- 支持自定义函数，提供丰富的函数调用方式，包括轻量级、快速的和隔离较好的；
- Support user-defined functions and provide rich function call methods, including lightweight, fast and well isolated;
&nbsp;<br>

- 支持动态加载函数；
- Support dynamic loading function;
&nbsp;<br>

- 支持动态加载模块代码并执行；
- Support dynamic loading and execution of module code;
&nbsp;<br>

- 支持并发编程；
- Support concurrent programming/thread;
&nbsp;<br>

- 支持地址引用，类似指针但受一定的保护；
- Support address reference, similar to pointer but protected to some extent;
&nbsp;<br>

- 支持编译成单独的可执行文件以便发布或者代码保护；
- Support compilation into separate executable files for release or code protection;
&nbsp;<br>

- 支持以系统服务的方式运行；
- Support running in system service mode;
&nbsp;<br>

- 内置网络服务器和微服务框架，可以直接以服务器模式运行；
- Built-in network/API server mode and micro-service framework are natively supported;
&nbsp;<br>

- 由于极简的语法结构和超轻量级的脚本运行引擎，因此可以很方便地移植到任意语言中，目前支持的Go、Java、JavaScript就是三种特点很不相同的语言，但都可以轻松实现谢语言的支持；
- Because of its extremely simple syntax structure and ultra-lightweight script running engine, it can be easily transplanted to any language. Currently, Go, Java and JavaScript are three languages with very different characteristics, but they can easily support Xielang;
&nbsp;<br>

&nbsp;<br>
下面是谢语言常见的欢迎程序代码：
The following is the common welcome program code of Xielang:

```go
pln `欢迎来到谢语言的世界！`
```

命令行上用下面的命令执行后可得结果如下：
The following results can be obtained after the command line is executed with the following command:

```cmd
D:\tmp>xie welcome.xie
欢迎来到谢语言的世界！

D:\tmp>


```

&nbsp;

下面是常见用于性能测试的斐波那契数列生成代码（fib.xie），使用了递归函数调用：
The following is the Fibonacci sequence generation code (fib. xie) commonly used for performance testing, which uses recursive function calls:

```go
// 用递归函数计算斐波那契数列
// 计算序列上第18个数
// cal Fibonacci numbers(the 18th) by recursive function

// 压栈一个整数18，表示计算第18个数
push int 18

// 调用标号:fib出的函数代码进行计算
call $drop :fib

pln $pop

exit

// 递归运算的斐波那契计算函数
:fib
    var $n1
    pop $n1

    < $push $n1 #i2

    if $pop :label1

    :else
        dec $n1
        push int $n1
        call $drop :fib

        dec $n1
        push int $n1
        call $drop :fib

        add $push $pop $pop

        ret

    :label1
        push $n1
        ret

```

将计算出斐波那契数列第18位数字，如下所示：
The 18th digit of Fibonacci sequence will be calculated as follows:

```shell
D:\tmp>xie fib.xie
2584

D:\tmp>

```

&nbsp;

#### 语言设计构思（Language design conception）

&nbsp;

谢语言的出现，最初是因为希望有一个能够嵌入在各种语言（初期考虑的语言主要是Go、Java、JavaScript、C/C++、C#、Swift等）内的轻量级脚本语言，能够支持在后端微服务中热加载修改的代码，要求语言的语法简单而又速度相对较快，但是能够充分发挥宿主语言的丰富库函数优势。后来逐渐发现谢语言也具备可以成为一门全栈语言的潜力，希望它最终能够达到。

Xielang is expected to be a lightweight scripting language that can be embedded in various languages (the languages considered at the initial stage are mainly Go, Java, JavaScript, C/C++, C #, Swift, etc.), which can support the hot loading and modification of code in back-end micro-services. The syntax of the language is simple and relatively fast, but it can give full play to the advantages of rich library functions of the host language. It seems that Xielang has the potential to become a full-stack language, and hope it will be eventually.

借鉴汇编语言的思路，谢语言引入了堆栈和寄存器等概念，也因此在某些功能的实现上会比一般的高级语言显得复杂一些，但从速度上（包括语法解析的速度）考虑，还是值得的。但要求开发者对堆栈、寄存器等概念做一些简单的了解。
Referring to the idea of assembly language, Xielang introduces the concepts of stack and register, which makes the implementation of some functions more complex than that of general high-level languages, but it is still worth considering from the speed (including the speed of syntax parsing). However, developers are required to have a simple understanding of the concepts of stack and register.

设计的原则包括：
The design principles include:

- 尽量减少语法分析的成本，因此拒绝复杂的语法结构，基本都以单行指令为主，只有多行字符串会占超过一行；
- Try to reduce the cost of syntax analysis, so reject complex syntax structures. Basically, single-line instructions are the main instructions. Only multi-line strings will occupy more than one line;
&nbsp;<br>

- 不做标准库，只做内置指令集，保证语言短小精悍，功能能够支持一般而言80%以上的常见开发需求（其余功能可以从源码自行扩充）；
- Do not build a standard library, but only build a built-in instruction set to ensure that the language is short and concise, and the functions can support more than 80% of common development requirements in general (other functions can be expanded from the source code);
&nbsp;<br>

- 在精简指令集基础上，支持函数和外部函数，支持外部模块的引入，以保证功能扩充的可能性；
- On the basis of the reduced instruction set, it supports functions and external functions, and supports the introduction of external modules to ensure the possibility of function expansion;
&nbsp;<br>

- 要支持并发编程（虽然很多脚本语言不支持并发编程）；
- Support concurrent programming (although most scripting languages do not support concurrent programming);
&nbsp;<br>

- 面向对象编程属于较低的优先级，甚至可以不实现；
- Object-oriented programming belongs to a lower priority, and may not even be implemented;
&nbsp;<br>

谢语言还在积极开发中，欢迎提出各种建议。
Xielang is still under active development and welcome to put forward various suggestions.

&nbsp;

#### 安装教程（Installation）

&nbsp;

1.  直接在[官网](http://xie.topget.org/)下载最新的谢语言可执行文件或压缩包，然后将其放在某个目录下，最好在系统路径之内，如果下载的是压缩包则先将其解压。然后即可使用；
- Download the latest Xielang executable file or compressed package from the [official website](http://xie.topget.org/), and then put it in a directory, preferably within the system path. If you download a compressed package, decompress it first. Then it is ready to use;
   
2.  谢语言代码执行的一般方法是在命令行执行（确保谢语言的主程序在路径中，否则需要加上路径）：
- The general way to execute Xielang code is to execute it on the command line (ensure that the main program of Xielang is in the path, otherwise the path needs to be added in the command-line):

```shell
xie hello.xie
```
3. 谢语言的代码文件一般以“.xie”作为扩展名，但这并不是强制的。注意，由于操作系统的限制，扩展名前的“.”只能是英文的小数点；\
- Code files in Xielang generally use ". xie" as the extension, but this is not mandatory.

3. 谢语言的代码文件内部都是纯文本的格式，并且要求使用UTF-8编码；另外，为便于跨平台使用，不建议使用BOM头；
- The code files of Xielang are in plain text format, and UTF-8 encoding is required; In addition, in order to facilitate cross-platform use, BOM headers are not recommended;

4. 安装后可以使用下述命令行验证是否安装成功，并且路径设置正常：
- After installation, you can use the following command line to verify whether the installation is successful and the path setting is normal:

```shell
xie -example hello.xie
```

如果看到类似下面的输出，说明安装成功，并且开发环境准备就绪。
If you see output similar to the following, the installation is successful and the development environment is ready.

```shell
D:\tmpx>xie -example hello.xie
Hello world!

D:\tmpx>

```

6. 直接不带任何参数运行谢语言主程序，将会进入谢语言的交互式命令行环境，在这里可以直接输入一行一行的命令，然后可以立即得到反馈结果：
- Run the main program of Xielang directly without any parameters, and you will enter the interactive command line environment of Xielang, where you can directly enter the command line by line, and then you can immediately get the feedback results:

```shell
C:\Users\Administrator# xie
> version
> pln $pop
0.0.1
>       
```

交互式命令行程序可以用于快速测试一些语句，或进行简单的编程获取结果。
The interactive command line program can be used to quickly test some statements, or conduct simple programming to obtain results.

7. 另外，可以用-version参数查看当前谢语言的版本号：
- In addition, you can use the - version parameter to view the version number of the current Xielang:

```shell
D:\tmpx>xie -version
Xielang(谢语言) Version(版本) 1.0.5

D:\tmpx>

```

在交互式命令行环境中，可以用version指令查看版本：
In the interactive command line environment, you can use the version directive to view the version:

```shell
> version $pln
0.2.3
> 
```

&nbsp;

#### 代码编辑器（Code Editor）

&nbsp;

1.  谢语言的代码编辑器推荐Visual Studio Code或Notepad 3，都是免费的编辑器；也可以使用任何支持UTF-8编码的文本编辑器；
- Visual Studio Code or Notepad 3 are recommended by Xielang's code editor, which are free editors; You can also use any text editor that supports UTF-8 encoding;

2.  语法高亮方案建议选择Rust或Shell Script（Shell脚本即可），也可以选用Go、C语言等的语法高亮方案，目前还没有专属的；
- The syntax highlighting scheme is recommended to select Rust, Shell Script (shell script is enough), or go, C language and other syntax highlighting schemes, which are not exclusive at present;

3.  谢语言也内置了简单的图形化命令行编辑器和的文本行编辑器，简单的代码也可以用它们来编写；
- Xielang also has built-in simple graphical command line editor and text line editor, and simple code can also be written with them;

&nbsp;

#### 运行和查看例子代码（Run and view sample code）

&nbsp;

谢语言提供各种例子代码，可以在命令行中加上-example参数直接运行，例如上述斐波那契数列代码就可以直接用下面的命令行运行：
Xielang provides various example codes, which can be run directly with the -example parameter in the command line. For example, the Fibonacci sequence code can be run directly with the following command line:

```shell
xie -example fib.xie
```

运行后结果类似：
The results after operation are similar:

```shell

D:\tmpx>xie -example fib.xie
2584

D:\tmpx>

```

如果需要查看例子代码，可以再加上-view参数，就可以看到：
If you need to view the example code, you can add the -view parameter to get it:

```shell
D:\tmpx>xie -example -view fib.xie
// 用递归函数计算斐波那契数列
// 计算序列上第18个数
// cal Fibonacci numbers(the 18th) by recursive function

// 压栈一个整数18，表示计算第18个数
push int 18

// 调用标号:fib出的函数代码进行计算
call $drop :fib

...

    :label1
        push $n1
        ret

D:\tmpx>
```

当然，也可以用“>”等转向符将其输出到其他文件中：
Of course, you can also use ">" and other steering symbols to output it to other files:

```
xie -example -view fib.xie > d:\test\new.xie
```

因此，如果我们后面说道：“请参看例子代码test1.xie”，那么就意味着可以通过
Therefore, if we say later in this document, "Please see the example code test1.xie", it means that you can use the

```
xie -example -view test1.xie
```

这样的命令行来参看所述的例子代码。
for such a command line, to see the example code described.

&nbsp;

#### 快速入门及语言要点（Quick Tour and Language Essentials）

&nbsp;

##### - **基本语法**（Basic grammar）

&nbsp;

  * 作为一门脚本语言，我们设计的初衷是尽可能降低解析语法的开销，因此谢语言选用了类似命令行的语法：一般命令（或者也称作指令或语句，英文为instruction，常用简称instr）都是一行，单条指令中的元素之间用空格作为分隔，第一个元素叫做指令名，后面的都叫做指令参数。也就是说，每条指令都由一个指令名和若干个指令参数组成，当然，有些指令也可以没有任何参数。例如：
  * As a script language, our original intention is to reduce the overhead of parsing syntax as much as possible, so Xielang uses a syntax similar to the command line: general commands (or also called instructions or statements, sometimes "instr") are all one line, and the elements in a single instruction are separated by spaces, the first element is called the instruction name, and the following are called the instruction parameters. In other words, each instruction consists of an instruction name and several instruction parameters. Of course, some instructions can also have no parameters. For example:

  ```
    assign $a "abc"
  ```

  其中的assign是指令名，后面的$a和"abc"都是指令参数。本条指令将把字符串"abc"赋值给变量a（注意不包括引号）。
  Where assign is the instruction name, and the following $a and "abc" are the instruction parameters. This instruction will assign the string "abc" to variable a (note that quotation marks are not included).

&nbsp;

  * 唯一不是一行的情况是多行字符串，谢语言中用成对的反引号字符“`”来包起多行字符串（实际上也可以包起非多行的字符串，会在一定的情况下有用），例如：
  * The only case that is not a single line is a multi-line string. In Xielang, a pair of backquote characters "`" are used to enclose multi-line strings (in fact, non-multi-line strings can also be enclosed, which will be useful under certain circumstances), for example:

```
    assign $a `abc
    123`
```

  这将把变量a赋值为多行字符串"abc\n123"；
  This will assign the variable **a** to the multiline string "abc\n123";

  &nbsp;

  * 命令头尾的空白都将被忽略，因此可以使用适当的缩进来提高代码的可读性。如果参数等值中含有不想省略的空格字符，需要用双引号或反引号括起整个字符串；
  * The blank space at the beginning and end of the command will be ignored, so you can use appropriate indentation to improve the readability of the code. If the parameter equivalent contains space characters that you do not want to omit, you need to enclose the entire string with double quotation marks or back quotation marks;
  * 所有代码中的指令名、变量、字符串、标号等都是大小写敏感的，也就是说如果仅有大小写不同的两个变量名，将被认为指的是不同的变量。
  * Instruction names, variables, strings, labels, etc. in all codes are case-sensitive. That is to say, if only two variable names with different case are different, they will be considered to refer to different variables.
  
&nbsp;

##### - **代码注释**（Comments）

&nbsp;

  谢语言中仅支持行注释，可以用“//”或“#”来引导注释，支持“#”是为了在文本编辑器中选用“Shell脚本”的语法高亮方案时，可以使用Ctrl+/组合键来切换改行是否注释。
  In Xielang, only line comments are supported. You can use "//" or "#" to guide the comments. The support of "#" is to select the syntax highlighting scheme of "shell script" in the text editor. You can use Ctrl+/key combination to switch whether the line change is commented or not.

&nbsp;

##### - **变量声明（定义）**（Variable declaration/definition）

  &nbsp;

  谢语言通常使用var命令（中文命令为“声明变量”）进行变量声明：
  In Xielang, we usually use the **var** command to declare variables:

  ```go
  var $a
  ```

  这将定义一个名字为a的变量，谢语言中，变量名前都要加“$”字符以示区别。定义变量时可以加第二个参数指定变量类型，此时其中的值为该类型的空值；不指定类型时，变量默认为nil（无类型的空值）如下所示：
  This will define a variable named a. In Xielang, the variable name should be preceded by the "\$" character to show the difference. When defining a variable, you can add a second parameter to specify the variable type. At this time, the value is the null value of the type; When no type is specified, the variable defaults to nil (null value without type) as follows:

  ```shell
    C:\Users\Administrator# xie
    > var $a
    > pln $a
    <nil>
    >        
  ```

  注意，pln命令类似于一般语言中的println函数，会将后面的变量和数值一个个输出到命令行，末尾再输出一个换行符。可以看到变量a中的值确实是nil。
  Note that the pln command is similar to the **println** function in general languages. It outputs the following variables and values to the command line one by one, and then outputs a newline character at the end. You can see that the value in variable a is really nil.

&nbsp;

  如果变量未定义就使用，会显示“未定义”字样，例如，下面的代码：
  If the variable is used without definition, the word "undefined" will be displayed, for example, the following code:

  ```go
    pln "a:" $a

    var $a

    pln "a:" $a

  ```

  运行后会显示如下结果：
  The following results will be displayed after running:

  ```
    a: undefined
    a: <nil>
  ```
    
  因为第一次输出时，变量a尚未被定义。
  Because the variable a has not been defined at the time of the first output.

&nbsp;

##### - **数值与类型**（Value and types）

&nbsp;

  * 谢语言中，常用的基本类型包括：bool/布尔、int/整数、float/浮点数（即小数）、str/字符串，其中整数和浮点数均为64位，还有byte/字节、rune/如痕（用于Unicode字符）。复杂类型在后面再介绍。使用plo命令，可以看到某个变量的类型和数值：
  * In Xielang, common basic types include: bool(boolean), int(integer), float(floating point number, i.e. decimal), str(string), where both integer and floating point number are 64 bits, and byte(byte), rune(for Unicode characters). Complex types will be introduced later. Use the **plo** command to see the type and value of a variable:

  ```shell
    C:\Users\Administrator# xie
    > var $a float
    > plo $a
    (float64)0
    > 
  ```

  可以看出，变量a被定义为float64即64位浮点数，并初始化为值0。
  It can be seen that the variable a is defined as float64, which is a 64-bit floating point number, and initialized to the value 0.

  谢语言中的变量的类型可以任意改变，意味着谢语言是一门“弱类型”的语言，而不像Go、C/C++、Java等“强类型”的语言那样，变量一旦声明后只能改变数值而不能改变类型。
  The type of variables in Xielang can be changed at will, which means that Xielang is a "weakly typed" language. Unlike "strongly typed" languages such as Go, C/C++, Java, etc., variables can only change the value but not the type once declared.

&nbsp;

##### - **给变量赋值**（Assignment）

&nbsp;

  谢语言中给变量赋值用的是assign/=（即选用assign或=都可以表示这条指令，后面都会用类似的写法）指令：
  In Xielang, assign/=/assignment is used to assign values to variables (that is, select assign or '=' to indicate this command(instruction), which will be written in a similar way later):

  ```go
    assign $a 123
  ```

  这条命令将把变量a赋值为123，注意，这是字符串“123”，而不是数字123，因为谢语言中默认数值都是字符串类型。
  This command will assign the value of variable a to 123. Note that this is the string "123", not the number 123, because the default value in Xielang is string type.
    
&nbsp;

##### - **指定赋值的类型**（Specify the type of assignment）

&nbsp;

  如果想把变量a赋值为一个整数，可以选用下面两种方法之一：
  If you want to assign variable a to an integer, you can choose one of the following two methods:

  ```go
    assign $a #i123
    assign $a int 123
  ```

  第一种方法是谢语言中对数值指定类型的方法，在数值前加上“#”号开头带一个指定的英语字母，可以限定数值的类型，对于基本类型，“#i”表示整数，“#f”表示浮点数，“#b”表示布尔数值（后跟true或false），“#s”表示字符串，“#y”表示字节，“#r”表示如痕。
  The first method is specifying the type of a number in Xielang. The number is preceded by a "#" sign with a specified English letter, which can limit the type of a number. For the basic type, "#i" represents an integer, "#f" represents a floating point number, "#b" represents a Boolean value (followed by true or false), "#s" represents a string, "#y" represents a byte, and "#r" represents a rune(32bit signed int).


   &nbsp; 

  第二种方法是在数值前再加一个指定数据类型的参数，可以是“int”、“float”、“bool”、“str”、“byte”、“rune”等。
  The second method is to add a parameter of the specified data type before the numerical value, which can be "int", "float", "bool", "str", "byte", "run", etc.

  看一下下面的输出，可以看到两种方法得到的结果是一样的：
  Looking at the following output, we can see that the results obtained by the two methods are the same:

  ```shell
    D:\tmpx# xie
    > assign $a #i123
    > plo $a
    (int)123
    > assign $a int 123
    > plo $a
    (int)123
    > = $变量1 整数 123
    > plo $变量1     
    (int)123
    >      
  ```

&nbsp;

##### - **字符串赋值**（String assignment）

&nbsp;

  由于谢语言使用空格作为命令与参数之间的分隔符，因此带有空格的字符串必须做特殊处理，使用双引号、单引号或反引号括起来（不含空格的字符串可以不括起来直接使用），双引号中可以带有\\n、\\t、\\"（表示双引号本身）等转义字符，单引号和反引号括起的字符串都不进行转义，反引号还可以括起多行字符串（含有换行符“\\n”的字符串）。另外，由于使用了反引号，谢语言代码中不应出现其他用途的反引号，如果遇上确需使用的地方，需要用全局变量\$backQuoteG或者转义字符“\u0096”来代替。

  Because Xielang uses spaces as the separator between commands and parameters, strings with spaces must be treated specially, using double quotation marks, single quotation marks or back quotation marks (strings without spaces can be used directly without being enclosed), and double quotation marks can contain \\n, \\t, \\"(indicating the double quotation mark itself) and other escape characters. The string enclosed by the single quotation mark and the back quotation mark cannot be escaped. The back quotation mark can also enclose multiple lines of string (containing the newline character"\\n"). In addition, because the back quotation mark is used, the back quotation mark for other purposes should not appear in Xielang code. If it is really necessary to use it, it needs to be replaced by the global variable \$backQuoteG or the escape character "\\u0096".



&nbsp;

##### - **各种赋值示例**（Examples of various assignments）

&nbsp;

  ```go
    assign $s abc
    plo $s

    assign $s "abc 123"
    plo $s

    assign $s `abc 123
    and this`
    plo $s

    assign $b int 3
    plo $b

    assign $b #i3
    plo $b

    assign $b #f3
    plo $b


  ```

  本段例子代码（assign.xie）执行后的结果是：
  
  The result of the execution of this example code (assign. xie) is:

  ```shell
    (string)abc
    (string)abc 123
    (string)abc 123
    and this
    (int)3
    (int)3
    (float64)3
  ```

&nbsp;

##### - **用var指令的时候赋值**（Assign value when using "var" instruction）

var指令在指定类型后面，也可以带有初始化赋值的数据，例如：

After the specified type, the **var** instruction can also carry data for initialization assignment, for example:

```go

var $a int 10

var $b string "abc非常好"

```

&nbsp;



##### - **堆栈**（Stack）

&nbsp;

堆栈是各种语言都会用到的数据结构，当然除了汇编语言外，一般都是“暗中”使用。但谢语言中将堆栈放开了使用，这有利于程序的性能，以及开发者灵活地操控。当然，对于对编程底层不是很了解的开发者来说，需要有一个适应的过程，容易犯错导致程序运行出乎意料。但熟悉之后，会发现这是一个很有力、很高效的编程基础设施。

Stack is a data structure used by all languages. Except for assembly language, of course, it is generally used "secretly". However, Xielang has released the use of the stack, which is conducive to the performance of the program and the flexible control of developers. Of course, for developers who don't know much about the underlying programming, there needs to be an adaptation process, which is easy to make mistakes and lead to unexpected program operation. But after getting familiar with it, you will find that it is a very powerful and efficient programming infrastructure.

&nbsp;

堆栈实质上是一个“后进先出”的队列，我们一般将其形象地想象为一个竖立的箱子，一般的操作包括“入栈”（英语为push，将一个数值压入堆栈，即放在堆栈的顶部）、“出栈”（英语为pop，将一个数值弹出堆栈，即从堆栈顶部取出一个数值）和“看栈”（英语为peek，即取到堆栈顶部的第一个数值，但并不做出栈操作，并不改变堆栈内容）。

The stack is essentially a "last in, first out" queue. We generally imagine it as an upright box. The general operations include "push" (push a value into the stack, that is, put it on the top of the stack), "Pop" and "peek" (look at the top of the stack).

&nbsp;

形象化地，我们有时候将入栈操作也称为“压入堆栈”，将出栈操作称为“弹出堆栈”、“弹出栈顶数值”，将看栈操作称为“查看栈顶”等。如果后面说道“弹栈值”，是指做了出栈（弹栈）操作后得到的值。另外，堆栈内的数值有可能被称为“元素”。

Visually, we sometimes call stack push operation as "push stack", stack pop operation as "pop stack", "pop stack top value", and stack view operation as "view stack top". If "pop stack value" is mentioned later, it refers to the value obtained after the pop stack operation. In addition, the values in the stack may be called "elements" or "items".

&nbsp;

堆栈在各种数值转移、计算、函数调用等场景中都发挥着重要的作用，谢语言中将其放在了明面上，给开发者提供一种高效的工具。

Stacks play an important role in various numerical transfer, calculation, function call and other scenarios. Xielang puts them on the bright side, providing developers with an efficient tool.

&nbsp;

##### - **基本的堆栈操作**（Basic stack operations）

&nbsp;

下面的代码（stack.xie）演示了堆栈的各种基本操作，代码中也写有详细的注释说明每条语句的作用，我们后面将大量使用这种方式来做代码示例和语法与指令讲解：

The following code (stack.xie) demonstrates various basic operations of the stack. Detailed comments are also written in the code to explain the function of each statement. We will use this method extensively to explain code examples, syntax and instructions later:

  ```go
 
  // 将整数2压入堆栈
  // push a integer value 2
  push #i2

  // 弹出栈顶数值到变量a中
  // pop the top item of the stack to variable $a
  pop $a

  // 输出变量a的内容
  // print the value of variable $a for reference
  plo $a

  // 将整数3压入堆栈
  // push another integer number 3 to stack
  push #i3

  // 将小数2.7压入堆栈，此时栈内从下而上包含两个元素：整数的3和浮点数2.8
  // push the decimal 2.7 onto the stack. At this time, the stack contains two elements from the bottom up: the integer 3 and the floating point number 2.8
  push #f2.8

  // 查看栈顶元素并将其赋值给变量b
  // view the top element of the stack and assign it to variable b
  peek $b

  // 输出变量b的内容
  // view the value of variable $b
  plo $b

  // 弹出栈顶元素到变量c中
  // pop the stack top element into variable c
  pop $c

  // 输出变量c的内容
  // view the value of variable $c
  plo $c

  ```

运行这段代码将输出：

After running this code, you will get output like below:

  ```shell
    (int)2
    (float64)2.8
    (float64)2.8
    (bool)true
    (bool)true
  ```

可以根据代码中的注释，详细观察堆栈操作的结果与预期的是否一致。

According to the comments in the code, you can observe in detail whether the result of the stack operation is consistent with the expected.

&nbsp;

##### - **与堆栈有关的特殊变量**（Special variables related to stack）

&nbsp;

谢语言中，有几个与堆栈操作有关特殊变量，属于系统预定义的变量，可以随时使用以便于一些灵活的数值操作。它们是“\$push”、“\$pop”和“\$peek”，分别表示入栈值、出栈值、看栈值。下面是它们的使用例子(stackVar.xie)：

In Xielang, there are several special 'global' variables related to stack operations, which are predefined variables in the system and can be used at any time to facilitate some flexible numerical operations. They are "\$push", "\$pop" and "\$peek", which represent the stack push operation, the stack pop value and the stack peek value respectively. The following are examples of the usage (stackVar.xie):


  ```go
  // 将字符串压入堆栈
  // push string to the stack
  push "我们高兴！"

  // 弹出栈顶数值，并输出
  // 注意弹出的数值如果不赋值给变量将丢失
  // pop up the value at the top of the stack and output
  // note that the pop-up value will be lost if it is not assigned to the variable
  plo $pop

  // 将整数18入栈
  // push an integer value 18 to the stack
  push #i18

  // 将出栈的数值赋值给变量a
  // assign the value on the top of the stack to the variable $a, and "drop" it(from the stack)
  assign $a $pop

  // 输出变量a
  // output variable a
  plo $a

  // 将浮点数3.14入栈
  // put floating point number 3.14 on the stack
  push #f3.14

  // 将栈顶值赋值给变量a
  // 此时堆栈内该数值仍将继续存在
  // assign the stack top value to variable $a
  // at this time(peek), the value will continue to exist in the stack
  assign $a $peek

  // 再次输出变量a
  // output variable $a again
  plo $a


  // 用assign语句将整数18入栈
  // $push变量表示将后面的数值压栈
  // use the assign statement to put the integer 18 on the stack
  // the $push variable means to push the following values on the stack
  assign $push #i3

  // 输出栈顶元素
  // output stack top element
  plo $peek
  ```

本段代码运行的结果是：

The result of running this code is:

  ```shell
    (string)我们高兴！
    (int)18
    (float64)3.14
    (int)3
  ```
   
&nbsp;

##### - **常见运算**（Common operations）

&nbsp;

先看看这个加法的例子（add.xie）：

First, take a look at the example of addition (add.xie):

  ```go
  // 将整数2入栈
  // push an integer value 2 to stack
  push #i2

  // 将整数5入栈
  // push another integer value 5 to stack
  push #i5

  // 将栈顶两个数值取出相加后结果压入栈中
  // 此处使用了预定义全局变量$push
  // 此时栈中应仅有一个数值（整数5）
  // add 2 values popped from the stack and add them
  // since we used the global predefined variable $push,
  // the result will be pushed into the (empty now) stack
  // after that, there is only one value 7 in the stack
  add $push $pop $pop

  // 输出栈顶数值（同时该数值被弹出）
  // output the top value of the stack
  plo $pop

  // 将浮点数1.5与2.6相加后压栈
  // add float value 1.5 and 2.6, push the result
  add $push #f1.5 #f2.6

  // 弹栈输出
  // print(and pop) the top value of the stack again
  plo $pop

  // 将两个字符串相加（连接）后赋值给变量c
  // add 2 string value(concat them) and put the result into variable $c
  add $c `abc` `123 456` 

  // 输出变量c
  // output variable $c
  plo $c

  // 将变量c中的数值压栈
  // push $c to the stack
  push $c

  // 将字符串“9.18”压栈
  // push a string "9.18" to the stack
  push "9.18"

  // 将栈顶两个字符串相加后赋值给变量d
  // Add the two strings at the top of the stack and assign the value to the variable $d
  add $d $pop $pop

  // 输出变量d
  // output variable $d
  plo $d

  // 将整数18与190相加后，压入栈中
  // Add the integers 18 and 190 and push them onto the stack
  add $push #i18 #i190

  // 弹栈输出
  // pop and output the result
  plo $pop
  ```

谢语言中，加法运算指令是add/+，将把结果参数之后的两个参数（可以是变量）值进行相加操作后将结果存入结果参数指明的变量中，如果没有指定结果变量，则存入全局预设变量\$tmp中。本例中我们使用了预置全局变量\$push表示将计算结果压入堆栈中。这段代码的运行结果是：

In Xielang, the addition operation instruction is add/+. After adding the values of the two parameters (which can be variables) after the result parameter, the result will be stored in the variable specified by the result parameter. If no result variable is specified, it will be stored in the global preset variable \$tmp. In this example, we use the preset global variable \$push to push the calculation result into the stack. The result of running this code is:

  ```shell
    (int)7
    (float64)4.1
    (string)abc123 456
    (string)abc123 4569.18
    (int)208
  ```

&nbsp;

其他类似的运算指令还有sub/-、mul/*、div//、mod/%等，用法类似。这些都属于二元运算指令，即参与运算的数值是两个。二元运算的两个数值必须是同一类型的。如果是不同类型，例如整数和浮点数相加，则需要先进行类型转换。

Other similar operation instructions include sub/-, mul/*, div//, mod/%, etc., with similar usage. These are binary operation instructions, that is, two values are involved in the operation. The two values of a binary operation must be of the same type. If it is of different types, such as adding integer and floating point numbers, type conversion is required first.

&nbsp;

##### - **数值类型转换**（Data type conversion）

&nbsp;

谢语言中，使用convert指令来转换数值类型，至少需要两个参数，第一个参数是数值或变量，第二个参数是字符串，指定需要转换成为的数据类型，如果有参数有三个，那么第一个参数（即结果参数）必须是一个变量，convert指令将会把转换后的结果存入该变量，否则会存入\$tmp。convert指令的使用示例（convert.xie）如下：

In Xielanguage, using the convert instruction to convert a numeric type requires at least two parameters. The first parameter is a numeric value or variable, and the second parameter is a string. Specify the data type to be converted into. If there are three parameters, the first parameter (that is, the result parameter) must be a variable, and the convert instruction will store the converted result in the variable, otherwise it will store it in  $tmp. An example of the use of the convert instruction (convert.xie) is as follows:

  ```go
  // 将整数15赋值给变量a
  // assign integer value to variable $a
  assign $a #i15

  // 此时如果执行指令 add $result $a #f3.6
  // 将会出现运行时错误
  // 应为加法运算的两个数值类型不一致
  // 一个是整数，一个是浮点数
  // at this time, if we execute the command "add $result $a # f3.6"
  // a runtime error will occur
  // the two numeric types expected for addition operation are inconsistent
  // one is an integer and the other is a floating point number

  // 输出变两个的数据类型和数值进行查看
  // pl指令相当于其他语言中的printf函数，后面再多输出一个换行符\n
  // Output the data type and value of two variables to view
  // pl instruction is equivalent to the printf function in other languages, followed by an additional newline character "\n"
  pl `a(%T)=%v` $a $a

  // 将变量a转换为浮点数类型
  // 结果将压入栈中
  // convert the variable $a to float point number
  // and push the result into the stack
  convert $push $a float

  // 输出栈顶值（不弹栈）的类型和数值查看
  // output the top value of the stack(no pop) for reference
  pl `a(%T)=%v` $peek $peek

  // 将栈顶值与浮点数3.6相加后压栈
  // pop the stack and add it with float point number 3.6
  // push the result to the stack
  add $push $pop #f3.6

  // 输出栈顶值查看类型和结果
  // 注意第一个参数使用$peek是为了不弹栈
  // 以保证第二个参数$pop操作时还能取到该值
  // output the stack top value to view the type and operation result
  // note that $peek is used for the first parameter to avoid stack pop action
  // to ensure that the value of the second parameter $pop can be obtained during operation
  pl "result=(%T)%v" $peek $pop

  ```

代码中解释很详细，运行结果如下：

The explanation in the code is very detailed, and the running results are as follows:

  ```shell
    a(int)=15
    a(float64)=15
    result=(float64)18.6
  ```

&nbsp;

##### - **字符串的连接操作**（String connection operation）

&nbsp;

谢语言中可以使用adds指令将多个字符串进行连接/拼接（有时候也叫字符串相加）。当然，adds指令不仅仅用于字符串的相加，也可以用于其他数据类型的相加，与add指令不同的是，adds指令可以将多个数值相加，并且可以用于不同类型的数值相加。adds指令会从左到右，将第一个数值与第二个数值相加，其结果再与第三个数值相加，依此类推直至加完所有数值。如果数值类型不同，adds指令将尽量把每次加法操作的第二个数值转换成第一个数值的类型，如果实在无法完成的加法，将返回error对象。

In Xielang, you can use the add instruction to connect/splice multiple strings (sometimes called string addition). Of course, the adds instruction is not only used for the addition of strings, but also for the addition of other data types. Unlike the add instruction, the adds instruction can add multiple values and can be used for the addition of different types of values. The add command will add the first value to the second value from left to right, and the result will be added to the third value, and so on until all the values are added. If the number types are different, the add instruction will try to convert the second number of each addition operation to the type of the first number. If the addition cannot be completed, the error object will be returned.

因此，对于相连接多个字符串的时候，或者想将包含字符串和数字等数值拼接成一个大字符串时，可以考虑使用adds指令。

Therefore, when connecting multiple strings, or when you want to splice values including strings and numbers into a large string, you can consider using the add command.

下面看一下例子（adds.xie）：

Let's take a look at the following example (adds.xie):

```go
// 本例演示adds指令的用法
// adds指令可以将多个数值相加，并且可以用于不同类型的数值相加
// adds指令会从左到右，将第一个数值与第二个数值相加，其结果再与第三个数值相加，依此类推直至加完所有数值
// 如果数值类型不同，adds指令将尽量把每次加法操作的第二个数值转换成第一个数值的类型
// 如果实在无法完成的加法，将返回error对象
// The add instruction can add multiple values and can be used to add different types of values
// The add command will add the first value to the second value from left to right, and the result will be added to the third value, and so on until all the values are added
// If the number types are different, the add instruction will try to convert the second number of each addition operation to the type of the first number
// If the addition cannot be completed, the error object will be returned

// 将多个字符串相加
// 注意其中含有一个浮点数3.8，将转换为字符串
// 另外，双引号、单引号，反引号都可以用于括起字符串，它们的区别是：
// 双引号括起的字符串可以包含转义字符，如\n、\"（表示双引号本身）等
// 单引号括起的字符串不进行转义
// 反引号支持多行字符串，括起的字符串也不进行转义
// Add multiple strings
// Note that it contains a floating point number 3.8, which will be converted to a string
// In addition, double quotation marks, single quotation marks and back quotation marks can be used to enclose strings. Their differences are:
// The string enclosed by double quotation marks can contain escape characters, such as  n,  "(indicating the double quotation marks themselves), etc
// Strings enclosed in single quotation marks are not escaped
// Backquotes support multi-line strings, and enclosed strings are not escaped
adds $result "abc" "\"123\"" #f3.8 '"递四方ds' `give it to 
    them
`

plo $result

// 进行依次整数相加，因为第一个数值$a是整数类型
// 因此后面的所有参数都将转换成整数再进行计算
// Perform sequential integer addition, because the first value $a is of integer type
// Therefore, all subsequent parameters will be converted to integers and then calculated
assign $a int 15

adds $result2 $a 30 #f2.3 #btrue

plo $result2
```

代码的执行结果是：

```shell
(string)"abc\"123\"3.8\"递四方dsgive it to \n    them\n"
(int)48
```

&nbsp;

##### - **指令的结果参数**（Result parameter of instruction）

&nbsp;

谢语言中大多数指令会产生一个或多个结果值（类似于其他语言中的函数会有返回值），谢语言中指令的返回值多数情况下是一个（函数的返回值视情况会有0个、1个或多个），也有少数指令有多个结果值。

Most instructions in Xielang will produce one or more result values (similar to the return values of functions in other languages). The return values of instructions in Xielang are mostly one (the return values of functions may be 0, 1 or more depending on the situation), and a few instructions have multiple result values.

因此，很多指令需要的一个用于指定接收指令执行结果的参数，我们将其称作结果参数。结果参数一般都是一个变量，因此也称作结果变量。结果变量可以是\$push（表示将结果压入堆栈中）、\$drop（表示将结果丢弃）等预置全局变量。结果变量有时可以省略，此时表示将结果存入全局变量\$tmp中（等同于显式声明为\$tmp）。但当指令的参数个数可变时，结果参数不可省略，以免产生混淆。因此，为清晰起见，一般情况下建议尽量显式使用结果参数。

Therefore, many instructions need a parameter to specify the execution result of the received instruction. We call it "the result parameter"(sometime the RP). The result parameter is generally a variable, so it is also called a result variable(i.e. RV). The result variables can be preset global variables such as \$push (meaning to push the result into the stack), \$drop (meaning to discard the result), etc. The result variable can sometimes be omitted, which means that the result is stored in the global variable \$tmp (equivalent to explicitly declared as \$tmp). However, when the number of parameters of an instruction is variable, the resulting parameters cannot be omitted to avoid confusion. Therefore, for the sake of clarity, it is generally recommended to declare the result parameters explicitly as much as possible.

例如toUpper指令被用于将字符串转换为大写，toUpper "abc" 会将大写的ABC存入\$tmp中， 而 toUpper $result "abc" 则会将ABC赋值给变量result。

For example, the toUpper instruction is used to convert strings to uppercase. toUpper "abc" will store uppercase ABC in \$tmp, while toUpper \$result "abc" will assign ABC to variable \$result.

另外，如果指令应返回结果，则文档中当不提结果参数时，“第一个参数”一般指的是除结果参数外的第一个参数，余者类推。

In addition, if the instruction should return a result, when the result parameter is not mentioned in the document, the "first parameter" generally refers to the first parameter except the result parameter, and so on.

对于带有可选个数参数的指令，则一般第一个参数必须是结果变量，不可省略，这样最后才可以接可选的n个参数，否则容易产生混淆。例如getWeb指令，一个典型用法（参看httpClient.xie）是：

For instructions with optional number of parameters, generally the first parameter must be the result variable and cannot be omitted, so that the optional n parameters can be connected at the end, otherwise it is easy to cause confusion. For example, a typical use of the getWeb instruction (see httpClient.xie) is:

```go

getWeb $resultT "http://127.0.0.1:80/xms/xmsApi" -method=POST -encoding=UTF-8 -timeout=15 -headers=`{"Content-Type": "application/json"}` $mapT

```

因为后面的参数除了URL是必须的外，其他都是可选的，不能确定有几个参数，因此只能把结果变量放在第一个，以便存放获取到的HTTP响应的内容。

Because the following parameters are optional except the URL, and it is uncertain how many parameters there are. Therefore, the result variable can only be placed in the first place to store the content of the obtained HTTP response.
   
&nbsp;

##### - **pl指令**（ the "pl" instr）

&nbsp;

上例中用到的pl指令，类似于一般语言中的printf函数，可以用占位符来控制输出的字符串内容。参数中第一个是格式字符串，可以含有%d、%f、%s、%v等占位符表示不同的数值输出形式，具体请参考Go语言等的参考文档。pln、plo、pl等指令在调试中经常会使用到，需要熟悉。

The pl instruction used in the above example is similar to the printf function in general languages. It can use placeholders to control the output string content. The first of the parameters is a format string, which can contain placeholders such as% d,% f,% s,% v, etc. to represent different numerical output forms. For details, please refer to the reference documents of Go language. Pln, plo, pl and other instructions are often used in debugging and need to be familiar with.

&nbsp;

##### - **内置全局变量**（Predefined/built-in global variables）

&nbsp;

我们前面已经接触到了一些谢语言中常用的内置全局变量，例如$push，$pop，$peek等，这里再列出所有的全局变量作为参考。

We have seen some built-in global variables commonly used in Xielang, such as \$push, \$pop, \$peek, etc. Here we will list all the predefined global variables for reference.

- **\$tmp** 
  表示内置全局变量tmp，例如 add #i1 #i2，将把整数1加2的结果存入\$tmp中，我们也可以像普通变量一样使用\$tmp。注意：tmp中存储的值我们在后面常简称做tmp值或者直接称作\$tmp。另外，使用\$tmp变量的原则是尽量快，因为任何指令或者表达式的计算都可能用到\$tmp，并令其发生改变。
  Indicates that the built-in global variable tmp, such as 'add #i1 #i2', will store the result of integer 1+2 in \$tmp. We can also use \$tmp as a common variable. Note: The value stored in \$tmp is often referred to as tmp value or \$tmp directly. In addition, the principle of using the \$tmp variable is to be as fast as possible, because any instruction or expression calculation may use \$tmp and make it change.

- **\$push** 
  表示压栈，例如 "add $push #i1 #i2"，将把整数1加2的结果压栈
  Indicates stack pushing. For example, "add \$push #i1 #i2" will push the result of integer 1 plus 2

- **\$pop** 
  表示弹栈，例如 "add $push $pop #i3"，将把弹栈值加上整数3，然后结果压栈
  Indicates the stack popping. For example, "add $push $pop #i3" will add the pop stack value to the integer 3, and then the result will be pushed.
  
- **\$peek** 
  表示看栈（不弹栈） ，用法类似$pop，但不弹出栈顶值（即保留在栈顶），而只是获取其值供使用
  It means to look at the stack (without popping the stack). Its usage is similar to \$pop, but it does not pop up the value at the top of the stack (that is, keep it at the top of the stack), but just gets its value for use
  
- **\$pln** 
  表示输出，例如：md5 $pln abc，将把字符串abc的MD5编码输出到命令行（类似于println函数的方式，结尾会输出一个回车符）
  Indicates the output, for example: "md5 $pln abc". The MD5 encoding of the string abc will be output to the command line (similar to the println function, with a carriage return at the end)

- **\$drop** 或/or **\$_**
  表示丢弃，通常在不关心指令执行结果时使用，例如：removeFile \$drop "c:\temp\tmp.txt"，将删除相应文件后，将执行结果丢弃
  Indicates discarding, which is usually used when the instruction execution result is not concerned, for example: removeFile \$drop "c:\\temp\\tmp.txt". After the corresponding file is deleted, the execution result will be discarded
  
- **\$seq** 
  表示一个全局的整数，每次使用都会加1，一般用于获取自增长、不重复的序号
  Represents a global integer, which will be increased by 1 every time it is used. It is generally used to obtain self-growing and non-repeating serial numbers
  
- **\$flexEvalEnvG** 
  用于灵活表达式做参数时的计算参数变量
  Used for Arguments While Flexible Expressions as Parameters
  
- **\$undefinedG** 
  表示未定义的变量值，或指令应返回结果而没有返回结果的时候
  Indicates an undefined variable value, or when the instruction should return a result without returning a result
  
- **\$debug** 
  表示获取当前的调试信息
  Indicates to get the current debugging information
  
- **\$argsG** 
  一般是指命令行参数，字符串列表类型
  It generally refers to the command line parameter, string list type
  
- **\$inputG** 
  一般是外部传入虚拟机的参数，可以是任意类型
  Generally, it is the parameter passed into the virtual machine from the outside, which can be any type
  
- **\$outG** 
  一般用这个变量保存虚拟机结束执行时像外部传出的参数（即返回值），可以是任意类型
  Generally, this variable is used to save the parameters (return value) that are sent out from the outside when the virtual machine ends execution, which can be of any type
  
- **\$paraMapG** 
  一般在HTTP请求响应模块中，表示请求参数（包括URL参数即GET/QUERY参数和POST参数）
  Generally, in the HTTP request response module, it represents the request parameters (including URL parameters, namely GET/QUERY parameters and POST parameters)
  
- **\$requestG** 
  一般在HTTP请求响应模块中，表示请求对象
  Generally, in the HTTP request response module, it represents the request object

- **\$responseG** 
  一般在HTTP请求响应模块中，表示响应对象
  Generally, in the HTTP request response module, it represents the response object
  
- **\$backQuoteG** 
  表示反引号字符
  Represents a backquote character
  
- **\$newLineG** 
  表示换行字符（即\\n）
  Indicates line feed character(i.e. \\n)
  
- **\$scriptPathG** 
  表示当前执行的脚本所在路径
  Indicates the path of the currently executed script
  
- **\$guiG** 
  GUI编程中的全局引用对象
  Global reference objects in GUI programming
  
注意，要避免自定义变量与这些变量名称冲突。
Note that you should avoid conflicts between custom variables and their names.

&nbsp;

##### - **标号**（Labels）

&nbsp;

谢语言中，可以在任意代码行的前一行添加标号，主要用于各种循环和条件分支等跳转场景。设置标号必须单独占一行，并以冒号“:”字符开头。

In Xielang, you can add a label to the previous line of any code line, which is mainly used for jump scenarios such as loops and conditional branches. The setting label must occupy a separate line and begin with a colon ":" character.

  ```go
    :lable1
    pln 123
  ```
   
&nbsp;

##### - **代码缩进**（Code Indent）

&nbsp;

谢语言中，每行代码的头尾空白将被忽略，因此可以适当采用代码的逐级缩进来增加代码的可读性。

In Xielang, the blank space at the beginning and end of each line of code will be ignored, so the progressive indentation of the code can be appropriately used to increase the readability of the code.


  ```go
    :lable1
        pln 123
  ```

   
&nbsp;

##### - **复杂表达式分解**（Complex expression decomposition）

&nbsp;

谢语言中，由于采用接近汇编语言的快捷语法，因此在一般的计算上或许要稍微复杂一些。通常情况下，建议对多步运算表达式采用分解后逐个进行的方式。例如，一个3+(9*1.5)/1.7的算式，建议用下面的代码（expression.xie）实现：

In Xielang, due to its fast syntax close to assembly language, it may be slightly more complicated in general calculation. Generally, it is recommended to decompose the multi-step operation expression one by one. For example, a 3+(9 * 1.5)/1.7 formula is suggested to be implemented with the following code (expression.xie):

  ```go
  // 计算3+(9*1.5)/1.7
  // Calculate 3+(9 * 1.5)/1.7

  // 将浮点数9压栈
  // push floating point number 9 onto the stack
  push #f9

  // 将浮点数1.5压栈
  // then push floating point number 1.5 onto the stack
  push #f1.5

  // 将栈顶两元素弹出相乘后结果存入预设全局变量tmp
  // store the result of multiplying the two elements at the top of the stack into the preset global variable tmp
  mul $pop $pop

  // 将tmp中的值和浮点数1.7相除后再次存入tmp
  // divide the value in tmp and floating point number 1.7 and store it in tmp again
  div $tmp #f1.7

  // 将浮点数3和tmp中值相加后存入$tmp
  // add the floating point number 3 and the value of tmp and save it into $tmp
  add #f3 $tmp

  // 输出结果查看
  // view output results
  pl "3+(9*1.5)/1.7=%v" $tmp

  ```

运行结果如下：

The operation results are as follows:

  ```
    3+(9*1.5)/1.7=10.941176470588236
  ```

可以看出，分解表达式的方法代码量比一般的高级语言多一些，但带来的好处是速度更快，因为省去了各种解析表达式的开销。后面可以看到，谢语言实际上也支持复杂的表达式运算，但显然自行分解的表达式运算效率更高。

It can be seen that the method code for decomposing expressions is a little more than that of general high-level languages, but the advantage is that it is faster because it saves the cost of various parsing expressions. As you can see later, Xielang actually supports complex expression operations, but it is obviously more efficient to decompose expressions by itself.
   
&nbsp;

##### - **复杂表达式运算**（Complex expression operation）

&nbsp;

谢语言中，也可以进行复杂的表达式计算，这要用到eval指令，参看下面的代码（eval.xie）：

In Xielang, complex expression calculation can also be performed, which requires the eval instruction. See the following code (eval. xie):

```go
// 本例演示表达式的使用
// This example demonstrates the use of expressions

// 给变量a赋值为整数12
// Assign the value of variable a to integer 12
assign $a #i12

// 计算表达式 a+(a+12/2) 的值，结果存入tmp
// 表达式是一个字符串类型的数值或变量
// 注意，一般的表达式有可能存在空格，因此需要用反引号或双引号括起来
// Calculate the value of expression a+(a+12/2) and store the result in tmp
// Expression is a numeric value or variable of string type
// Note that common expressions may have spaces, so you need to use back quotes or double quotes
eval "$a + ( $a + #i12 / #i2 )"

// 输出tmp值查看
// Output tmp value to view
pln $tmp

// 将变量b赋值为整数-9
// Assign variable b to integer - 9
assign $b #i-9

// 计算顺序括号优先，无括号时按照一般的运算符顺序进行计算
// 结果值放入变量r
// 本例要计算的表达式的数学表达是 a+((a-8.0)*abs(b))，其中abs表示取绝对值
// 注意由于计算顺序问题，数学表达中需要把a-8.0加上括号以保证计算顺序一致
// 表达式里可以包含指令，此时应该使用花括号将其括起来
// 该指令必须通过$tmp变量返回一个结果值继续参加表达式的运算，这样可以使得表达式中实现基本运算符之外的运算功能，例如转换数值类型等
// 花括号不可以嵌套使用
//The calculation order takes precedence over parentheses. If there are no parentheses, the calculation is performed according to the general operator order
//The result value is put into the variable r
//The mathematical expression of the expression to be calculated in this example is a+((a-8.0) * abs (b)), where abs represents the absolute value
//Note that due to the calculation order problem, it is necessary to add brackets to a-8.0 in the mathematical expression to ensure the consistent calculation order
//Expressions can contain instructions, which should be enclosed by curly braces
//The instruction must return a result value through the $tmp variable to continue to participate in the operation of the expression, which can enable the expression to implement the operation functions other than the basic operator, such as converting the numeric type, etc
//Curly brackets cannot be nested
eval $r `$a + ($a - {convert #f8.0 int}) * {abs $b}`

// 输出变量r的值查看
// View the value of output variable r
pln $r

// 判断表达式 !((a-b)<10) 的计算结果值是否为布尔值true，是则跳转到标号next1处
// ifEval指令后第一个参数必须是一个字符串类型的数值或变量，表示要计算的表达式
// 第二个参数时满足条件后要跳转到的标号
// Judge expression! Whether the calculated result value of ((a-b)<10) is a boolean value true, and if yes, it will jump to the label next1
// The first parameter after the ifEval instruction must be a numeric value or variable of string type, representing the expression to be evaluated
// The second parameter is the label to jump to when the condition is met
ifEval `! (($a - $b) < #i10)` :next1

pln 条件不满足
exit

:next1
pln 条件满足
```

需要特别注意的是，谢语言中的表达式中，运算符是没有优先级之分的，因此一个表达式中是严格按照从左到右的顺序执行运算的，唯一的例外是括号，用圆括号可以改变运算的优先级，括号里的部分将被优先计算。另外，表达式中的值与运算符之间必须有空格分隔。也因为一般的表达式都存在空格，因此需要用反引号或双引号括起来。

Special attention should be paid to the fact that in the expressions in Xielang, operators have no priority. Therefore, an expression performs operations in strict order from left to right. The only exception is parentheses. Parentheses can change the priority of operations, and the parts in parentheses will be calculated first. In addition, the value and operator in the expression must be separated by a space. Because there are spaces in general expressions, you need to enclose them with back quotes or double quotes.

另外，如果括号里的内容以一个问号“?”开始，那么后面可以是一条指令，该指令必须通过$tmp变量返回一个结果值以便继续参加表达式的运算，这样可以使得表达式中实现基本运算符之外的运算功能，例如转换数值类型等。

In addition, if the content in the parentheses starts with a question mark "?", then it can be followed by an instruction that must return a result value through the $tmp variable to continue to participate in the operation of the expression, which can enable the expression to implement the operation functions other than the basic operator, such as converting the numeric type.

ifEval指令是专门配合表达式计算使用的条件跳转指令，它后面必须跟一个字符串类型的表达式，其计算结果必须是一个布尔类型的值，ifEval指令将根据其结果，确定是否要跳转到指定的行号。ifEval指令，简化了一般的if和ifNot质量较为复杂的条件处理语法结构。

The ifEval instruction is a conditional jump instruction specially used for expression calculation. It must be followed by an expression of string type, and its calculation result must be a Boolean value. The ifEval instruction will determine whether to jump to the specified line number according to its result. The ifEval instruction simplifies the general if and ifNot quality conditional processing syntax structure.

由于谢语言中表达式计算相对效率较低，因此对于需要反复高速计算或处理的场景，建议还是使用分解的方式更高效。

Due to the relatively low efficiency of expression calculation in Xielang, it is recommended to use decomposition method for scenes that require repeated high-speed calculation or processing.

运行后的效果：

Effect after operation:

```shell
30
48
条件满足
```

&nbsp;

##### - **复杂表达式做参数**

&nbsp;

谢语言中，表达式可以运用在指令的参数中，此时需要以英文问号“?”字符开头，例如（exprInParam.xie）：

```go
// 本例演示指令中用表达式作为参数
// This example demonstrates using expressions as parameters in instructions

assign $a "abc"

// 表达式做参数
// 注意“@”后面再加双引号或反引号括起表达式
// Expression as parameter
// Note that the expression is enclosed by double quotation marks or back quotation marks after "@"
pl "[%v] test params: %v" @"{nowStr}" $a
```

将输出：

Will output:

```shell
[2022-05-17 14:30:59] test params: abc
```

其中，pl指令的第二个参数即是以@号开头的表达式，而这个表达式用花括号括起指令的方式又运行了获取当前时间字符串的指令nowStr。注意，表达式内的指令，一定要保证将结果值存入全局变量、$tmp（不可省略结果参数的指令，要确保结果参数是\$tmp）。

Among them, the second parameter of the pl instruction is the expression beginning with the @ sign, and this expression runs the instruction nowStr to obtain the current time string by enclosing the instruction with curly braces. Note that the instruction in the expression must ensure that the result value is stored in the global variable, \$tmp (for the instruction that cannot omit the result parameter, ensure that the result parameter is \$tmp).

&nbsp;

##### - **表达式的另一个例子**（Another example of an expression）

下例是另一个表达式的例子，使用quickEval指令，与eval指令是等价的（quickEval.xie）：

The following example is another example of an expression, using the quickEval instruction, which is equivalent to the eval instruction (quickEval.xie):

```go
// 本例展示快速表达式
// 注意快速表达式中需要用花括号来支持内嵌指令或函数
// This example shows a fast expression
// Note that curly braces are needed in fast expressions to support embedded instructions or functions

// 将变量a赋值为浮点数15.2
// Assign variable a to floating point 15.2
= $a #f15.2

// 计算 -5.1*2.8+(23+(a-6.9))/3.3
// quickEval指令用于计算一个用字符串表示的快速表达式的值
// Calculation - 5.1 * 2.8+(23+(a-6.9))/3.3
// The quickEval instruction is used to calculate the value of a fast expression expressed as a string
quickEval `-#f5.1*#f2.8+(#f23+ ($a -#f6.9)) /#f3.3 `

pln $tmp

// 计算 3+(16-2)/3%2 并输出结果
// Calculate 3+(16-2)/3% 2 and output the result
quickEval $pln `#i3 + (#i16 -#i2) / #i3 % #i2`

= $s1 "abc 12\n3 \u0022大家好\u0022"

// 计算字符串的相加（即连接）结果
// Calculate the result of adding (connecting) strings
quickEval $pln `" -- " + $s1 + "--"`

// 将变量b赋值为整数18
// Assign variable b to integer 18
assign $b #i18

// if指令后也可以接快速表达式表示判断条件
// 快速表达式做参数时，以@符号开始，一般后面用反引号括起来，因为常有空格
// if语句后快速表达式也可以不带@符号，直接是一个字符串，会自动判断
// The if instruction can also be followed by a fast expression to express the judgment condition
// When a fast expression is used as a parameter, it starts with the @ sign and is usually followed by a back quotation mark, because there are often spaces
// The quick expression after the if statement can also be a string without the @ sign, which will be automatically determined
if @`$b > #i12` +1 +3
    pl "$a > #i12"
    goto :next1

    pl "$a <= #i12"

:next1

// 给变量s1赋值为字符串abcde
// Assign the value of variable s1 to the string abcde
= $s1 `abcde`

// 快速表达式中如果需要进行内嵌指令运算，需要用花括号括起来
// 另外内嵌指令的结果必须存入临时变量$tmp中
// If the embedded instruction operation is required in the fast expression, it needs to be enclosed in curly brackets
// In addition, the result of the embedded instruction must be stored in the temporary variable $tmp
quickEval $rs `#i15*#i3+{toInt $tmp 19}* {len $tmp $s1}`

pl "first result: %v" $rs

plv @`#i15/#i3+{toInt $tmp 19}* {len $tmp $s1}-#i3`

// 内嵌指令中不能再使用花括号，其他值中可以使用花括号
// Curly brackets can no longer be used in embedded instructions, and can be used in other values
plv @`{toStr $tmp #i123456} + " {ab 123 c}"`

```

条件判断指令if中，可以直接带字符串类型的快速表达式，方便代码书写。

In the conditional judgment instruction if, you can directly take a string type of fast expression, which is convenient for code writing.

&nbsp;

##### - **灵活表达式**（Flex expression）

谢语言还支持接近其他语言中的语法的表达式，称作“灵活表达式”（Flex Expression），使用flexEval或flexEvalMap指令进行运算，具体用法请参看下面的例子（flexEval.xie）：

Xielang also supports expressions that are similar to the syntax in other languages, called "Flex Expressions", using the flexEval or flexEvalMap instructions for operations. For specific usage, please refer to the following example (flexEval.xie):

```go
// 本例介绍flexEval和flexEvalMap指令的各种主要用法
// This example introduces the various main uses of the flexEval and flexEvalMap instructions

// ----- 案例1：使用flexEval指令计算表达式，并传入多个参数
// Case 1: Using the flexEval instruction to evaluate an expression and passing in multiple parameters

// 赋值一个浮点数变量f1和一个整数变量c1
// Assign a Floating-point arithmetic variable f1 and an integer variable c1
= $f1 #f19.9
= $c1 #i23

// 使用flexEval指令计算两者的和
// flexEval指令用于计算较复杂的表达式，并可以进行自定义与扩展
// flexEval后的结果参数不可省略，然后第一个参数为字符串形式的表达式
// 表达式中可以用v1、v2等虚拟变量表示从第二个参数开始的各个值
// 因此这里是将变量f1和c1中的值进行求和
// 此时如果两个变量数据类型不同（例如这里一个是浮点数，一个是整数），会进行适当的转换
// Calculate the sum of the two using the flexEval instruction
// The flexEval instruction is used to calculate more complex expressions and can be customized and extended
// The result parameter after flexEval cannot be omitted, and the first parameter is an expression in string form
// Virtual variables such as v1 and v2 can be used in expressions to represent various values starting from the second parameter
// Therefore, here we sum the values in variables f1 and c1
// At this time, if the data types of two variables are different (for example, one is a Floating-point arithmetic number and the other is an integer), appropriate conversion will be performed
flexEval $result1 `v1 * v2 / 2` $f1 $c1

// 输出结果作为参考
// Output the result as a reference
pl "result1=%#v" $result1

// ----- 案例2：使用flexEvalMap指令计算表达式，并传入一个列表参数来代替多个参数
// Case 2: Using the flexEvalMap instruction to evaluate an expression and passing in a list parameter instead of multiple parameters

// 定义一个映射变量map1
// Define a mapping variable map1
var $map1 map

// 设置其中两个键值对，均为字符串类型
// Set two key value pairs, both of which are of string type
setMapItem $map1 "s1" "abc"
setMapItem $map1 "s2" "789"

// flexEvalMap与flexEval的不同是：flexEval后从第二个参数开始可以接受多个参数，并在表达式中以v1、v2这样来指代
// 而flexEvalMap则只允许有一个参数，需要是映射类型，这样可以直接用键名在表达式中引用这些变量
// The difference between flexEvalMap and flexEval is that after flexEval, it can accept multiple parameters starting from the second parameter and refer to them as v1 and v2 in the expression
// And flexEvalMap only allows one parameter, which needs to be a mapping type, so that these variables can be directly referenced in the expression using the key name
flexEvalMap $result2 `s1 + s2` $map1

// 输出结果作为参考
// Output the result as a reference
pl "result2=%#v" $result2

// ----- 案例3：表达式中使用自定义函数
// Case 3: Using Custom Functions in Expressions

// 定义一个实现简单字符串trim功能的函数dele1，注意这里是快速代理函数
// Define a function dele1 that implements the simple string trim function, and note that this is the fast proxy function
new $dele1 quickDelegate `
    getArrayItem $param1 $inputL 0

    trim $outL $param1

    exitL 
`

// 定义一个实现截断浮点数到小数点后2位功能的函数dele2，注意这里是普通代理函数
// Define a function dele2 that implements the function of truncating floating point numbers to 2 digits after the Decimal separator. Note that this is a general proxy function
new $dele2 delegate `
    getArrayItem $param1 $inputG 0

    spr $outG "%.02f" $param1

    exit 
`

// 定义一个映射变量map2
// Define a mapping variable map2
var $map2 map

// 设置其中几个键值对，包括列表，映射、自定义函数几种类型
// Set several key value pairs, including lists, mappings, and custom function types
setMapItem $map2 "a1" #L`[68, 37, 76]`
setMapItem $map2 "m1" #M`{"value1": -3.5, "value2": 9.2}`
setMapItem $map2 "trim" $dele1
setMapItem $map2 "toFixed" $dele2

// abs函数是表达式引擎内置的函数，功能是取绝对值，float也是内置函数，将字符串转换为浮点数
// trim和toFixed都是自定义函数
// 注意对于列表和映射使用的索引方法，与一般语言中的一样
// 计算结果应为19.4399
// The abs function is a built-in function of the expression engine, which takes absolute values. Float is also a Intrinsic function, which converts strings into Floating-point arithmetic numbers
// trim and toFixed are both custom functions
// Note that the indexing method used for lists and mappings is the same as in general languages
// The calculation result should be 19.4399
flexEvalMap $result3 `abs(float(trim("   "+toFixed(a1[0] / m1.value1)+"99  ")))` $map2

pl "result3=%#v" $result3

// ----- 案例4：表达式中使用缺省函数实现过滤和枚举
// Case 4: Using default functions in expressions to implement filtering and enumeration

// 赋值一个整数列表（数组）a2
// Assign an integer list (array) a2
= $a2 #L`[68, 37, 16, 9, 88, 76]`

// 将a2代入表达式中计算（注意在表达式中将由v1代表它）
// v1[1:5]表示对数组进行切片，取序号为1开始，到序号为5之前的数据项，其结果应为[37, 16, 9, 88]
// filter函数是表达式引擎内置函数，表示将某个数组以指定条件过滤，这里的条件是要求值大于20（“#”表示过滤时的每一项），过滤后结果应为[37, 88]
// map函数也是内置函数，表示将数组中每一项做一个操作后形成新的数组，这里是将每一项除以2，最终结果应为[18.5, 44]
// Substitute a2 into the expression for calculation (note that v1 will represent it in the expression)
// v1 [1:5] represents slicing an array, starting with sequence number 1 and ending with data items before sequence number 5. The result should be [37, 16, 9, 88]
// The filter function is the Intrinsic function of the expression engine, which means to filter an array with specified conditions. The condition here is that the value is required to be greater than 20 ("#" means each item during filtering). The filtered result should be [37, 88]
// The map function is also a Intrinsic function, which means that each item in the array is formed into a new array after an operation. Here, each item is divided by 2, and the final result should be [18.5, 44]
flexEval $result4 `map(filter(v1[1:5], {# > 20}), {# / 2})` $a2

// 以JSON格式输出变量result4中结果供参考
// Output the results in variable result4 in JSON format for reference
pl "result4=%#v" @`{toJson $tmp $result4}`


```

内置函数如果某些情况下出现问题，可以编写同名的自定义函数来代替。

If there is a problem with the Intrinsic function in some cases, you can write a user-defined function with the same name to replace it.

&nbsp;

##### - **灵活表达式做参数**（Flexible expression as parameter）

&nbsp;

灵活表达式也可以用在参数中，can看下面的例子（flexEval2.xie）：

Flexible expressions can also be used in parameters, as shown in the following example (flexEval2.xie):

```go

// 参数以@@开始表示一个灵活表达式，表达式一般需要用反引号括起来
// 在指令中的灵活表达式将被自动计算出结果值
// The parameter starts with @@ to represent a flexible expression, which usually needs to be enclosed in back quotes
// The flexible expression in the instruction will automatically calculate the result value
pl "result1=%#v" @@`int(7.1 * 24 / 88 - 6 * 99.3)`

// 定义一个实现简单字符串trim功能的函数dele1，注意这里是快速代理函数
// Define a function dele1 that implements the simple string trim function, and note that this is the fast proxy function
new $dele1 quickDelegate `
    getArrayItem $param1 $inputL 0

    trim $outL $param1

    exitL 
`

// 定义一个实现截断浮点数到小数点后2位功能的函数dele2，注意这里是普通代理函数
// Define a function dele2 that implements the function of truncating floating point numbers to 2 digits after the Decimal separator. Note that this is a general proxy function
new $dele2 delegate `
    getArrayItem $param1 $inputG 0

    spr $outG "%.02f" $param1

    exit 
`

// 定义一个dele3函数用于替代内置的abs函数
// Define a dele3 function to replace the built-in ABS function
new $dele3 quickDelegate `
    getArrayItem $param1 $inputL 0

    abs $outL $param1

    exitL
`

// 灵活表达式做参数时，将自动寻找全局预设变量flexEvalEnvG来获取所需的参数和自定义函数等
// When using flexible expressions as parameters, it will automatically search for the global preset variable flexEvalEnvG to obtain the required parameters and custom functions
var $flexEvalEnvG map

// 设置其中几个键值对，包括列表，映射、自定义函数几种类型
// Set several key value pairs, including lists, mappings, and custom function types
setMapItem $flexEvalEnvG "a1" #L`[68, 37, 76]`
setMapItem $flexEvalEnvG "m1" #M`{"value1": -3.5, "value2": 9.2}`
setMapItem $flexEvalEnvG "trim" $dele1
setMapItem $flexEvalEnvG "toFixed" $dele2
setMapItem $flexEvalEnvG "abs" $dele3

// toFixed和abs都是自定义函数
// 注意对于列表和映射使用的索引方法，与一般语言中的一样
// 计算结果应为19.43
// toFixed and abs are both custom functions
// Note that the indexing method used for lists and mappings is the same as in general languages
// The calculation result should be 19.43

pl "result2=%#v" @@`toFixed(a1[0] / abs(m1.value1))`

```

代码运行后将输出：

After running the code, it will output:

```shell
result1=-593
result2="19.43"
```

&nbsp;

##### - **goto语句**（The goto instr）

&nbsp;

goto语句在一般的高级语言中并不推荐使用，但对于具备一定经验的开发者来说，反而有可能是提高效率的手段。谢语言中提供了goto/转到指令（为保持和汇编语言的一定关联，也可写作jmp），可以用于实现代码执行中无条件跳转到某个标号处的功能。例如（goto.xie）：

The goto statement is not recommended in general high-level languages, but it may be a means to improve efficiency for developers with certain experience. Xielang provides goto/go instruction (in order to maintain a certain association with assembly language, it can also be written as jmp), which can be used to realize the function of unconditional jump to a label in code execution. For example (goto.xie):

  ```go
  pln start...

  push #f1.8

  goto :label1

  :label2
  pop $c

  pln `c =` $c

  exit

  :label1
      pln "label1 =" $peek

      goto :label2


  ```

由于无条件跳转的关系，这段代码执行时将先执行标号:lable1处的代码，然后再跳转到标号:label2处的代码，最后输出结果是：

Due to the unconditional jump, the code at label: lable1 will be executed first, and then jump to the code at label: label2. The final output result is:

  ```
  start...
  label1 = 1.8
  c = 1.8
  ```

另外，goto语句中的，可以使用“:+1”、“:-3”这种伪标号，表示跳转到当前指令的后一条指令或前三条指令（注意：注释和标号等不是有效指令的行将被忽略而不被计算在内）：

In addition, in the goto statement, pseudolabels such as ":+1" and ":-3" can be used to indicate jumping to the next instruction or the 3rd previous instructions of the current instruction (note: lines such as comments and labels that are not valid instructions will be ignored and not counted):

```go
  pln abc
  goto :+3

  // 下面两条指令将被跳过
  // the next two line of code will be skipped
  pln 123
  pass 

  pln "这句将被执行（this line of code will be run）"
```

&nbsp;

##### - **一般循环结构**（Loop cycle structure）

&nbsp;

循环结构是一般计算机语言中必然会有的基本语法结构。谢语言中，一般使用各种跳转语句来实现循环结构。goto语句是其中的一种方法，最常见的是实现无限循环。

Loop structure is the basic grammatical structure that is inevitable in general computer language. In Xielang, various jump statements are generally used to realize the loop structure. The goto statement is one of the methods. The most common method is to implement infinite loops.

  ```go
  // 将字符串压栈
  // infinite loop
  push "welcome"

  // 设定标号loop1
  // set label loop1
  :loop1
      // 输出栈顶值
      // pop the stack top value and output it
      pln $peek

      // 休眠2.5秒
      // sleep for 2.5 seconds
      sleep #f2.5

  // 跳转到标号loop1处继续往下执行
  // jump to the position at label :loop1
  goto :loop1
  ```

sleep指令的作用是休眠指定的秒数。本例（for1.xie）运行的结果是将每隔2.5秒输出一下“欢迎”两个字，直到按Ctrl-C等方法来终止程序的运行。

The sleep instruction sleeps for the specified number of seconds. The result of this example (for1.xie) is that the word "Welcome" will be output every 2.5 seconds until the program is terminated by pressing Ctrl-C and other methods.
   
&nbsp;

##### - **条件分支**（Conditional branch）

&nbsp;

谢语言中的条件分支支持一般是由比较判断指令和条件跳转指令结合来实现的。直接看下面的例子（if.xie）：

The conditional branch support in Xielang is generally realized by the combination of comparison and judgment instructions and conditional jump instructions. Look directly at the following example (if.xie):

  ```go
  // 给变量i赋值整数11
  // assign integer value 11 to variable $i
  assign $i #i11

  // 比较变量i是否大于整数10
  // 结果放入变量a中
  // compare if $i > 10(integer)
  // then put the result to $a
  > $a $i #i10

  // 判断$a是否为布尔值true
  // 如果是则跳转到标号label2
  // check if $a == true(bool value)
  // if true jump to :label2(label)
  if $a :label2
      // 否则执行下面的语句
      // if not met, continue to run the following
      pln "else branch"

  //终止程序执行
  // terminate the program
  exit

  // 标号label2
  // label named label2
  :label2
      // 输出“if branch”
      // output "if branch" for reference
      pln "if branch"

      // 将局部变量b赋值为整数8
      // assign a local variable $b(since there are no variabes with this name in global context) to integer value 8
      assign $b #i8

      // 比较变量b是否小于或等于变量i
      // 由于省略了结果变量，结果将被放入$tmp中
      // check if $b <= $i
      // the result variable is omitted, so the result will be put into global variable $tmp
      <= $b $i

      // 判断否（tmp值是false）则跳转到标号label3
      // ifNot指令是判断条件为false则跳转
      // check if $tmp is not true
      // if true($tmp is false), jump to label3
      ifNot $tmp :label3
          // 否则输出
          // else branch
          pln "label3 else"

      // 终止代码执行
      // terminate
      exit

      // 标号label3
      :label3
          // 输出“label3 if”
          pln "label3 if"
  ```

其中，出现了两个比较指令：“>”和“<=”，这些比较指令所带参数都和二元运算指令类似，可以从堆栈中取两个值做比较，也可以对后面所带的两个参数进行比较，当然还可以带一个参数（放在第一个）表示将结果赋值给某个变量，否则会将结果存入\$tmp。比较指令返回的结果都是布尔值true或者false。

Among them, there are two comparison instructions: ">" and "<=". The parameters of these comparison instructions are similar to the binary operation instructions. You can take two values from the stack for comparison, and you can also compare the following two parameters. Of course, you can also take a parameter (put in the first) to assign the result to a variable, otherwise the result will be stored in \$tmp. The results returned by the comparison instruction are boolean values of true or false.

&nbsp;

而条件跳转指令if和ifNot可以带1或2个参数，最后一个参数都是符合条件要跳转到的标号，如果还有第一个参数则表明要判断的变量或数值（必须是布尔值），没有的话则从堆栈取数进行判断：if指令是true则跳转，ifNot是false则跳转。

The conditional jump instructions if and ifNot can take 1 or 2 parameters. The last parameter is the label to jump to if the condition is met. If there is the first parameter, it indicates the variable or value to be judged (must be a boolean value). If there is no parameter, it will be judged from the stack fetch: if the instruction is true, it will jump, if not, it will jump.

&nbsp;

这段代码的运行结果是：

The result of running this code is:

  ```
    if branch
    label3 else
  ```

注意观察条件分支的流转是否符合预期。

Observe whether the flow of conditional branches meets expectations.

比较指令主要包括：==（等于）、!=（不等于）、>、<、>=、<=等。

The comparison instructions mainly include: ==(equal to), !=(not equal to), >, <, >=, <=, etc.

&nbsp;

##### - **else分支**（Else branch）

&nbsp;

if、ifNot等条件分支指令其实还支持第三个参数，即else分支。该参数也是一个标号，表示条件不满足时要走的分支。直接看下面的例子（else.xie）：

Conditional branch instructions such as if and ifNot actually support the third parameter, namely else branch. This parameter is also a label, indicating the branch to take when the condition is not met. Look directly at the following example (else.xie):

```go
> #i3 #i2

if $tmp :label1 :else1

:label1
    pln label1

    goto :next1

:else1
   pln else1 

:next1

> $push #f1.5 #f3.6
if $pop :label2 :else2

:label2
    pln label2

    exit

:else2
   pln else2


```

运行输出结果为：

The program output result is:

```shell
label1
else2
```

注意其中是否走了else分支。

Note whether else branch is taken.


&nbsp;

##### - **虚拟标号/伪标号跳转**（Virtual label/pseudolabel jump）

在无条件跳转指令goto和条件跳转指令if、ifNot、ifEval等语句中，不一定非要用标号表示跳转目的地，也可以使用类似“+1”，“+3”这种伪标号（更推荐写作:+1，:+3这样的形式，因为在fastCall调用的快速函数代码中，+1这种形式有可能失效，而:+1不会失效，另外:+1这种形式可以用在更广泛的地方，基本上所有用到标号的地方都可以使用），表示跳转到当前指令的后一条指令或前三条指令等。注意，注释和标号等不是有效指令的行将被忽略而不被计算在内。我们来看下面的例子（quickIf.xie）：

In the unconditional jump instruction goto and conditional jump instruction if, ifNot, ifEval and other statements, it is not necessary to use a label to indicate the jump destination, but it can also use a similar "+1", "+3" is a pseudolabel (it is more recommended to write as :+1, :+3, because in the fast function code called by FastCall, +1 may be invalid, but :+1 will not be invalid. In addition, :+1 can be used in a wider range of places, and can be used in almost all places where the label is used), indicating that the next instruction or the first three instructions of the current instruction can be skipped. Note that lines that are not valid instructions such as comments and labels will be ignored and not counted. Let's look at the following example (quickIf.xie):

```go
// 本例演示了在if和goto等指令中使用“+1”、“+3”等“伪标号”进行跳转的方法
// +1是指跳转到当前指令的下一条指令，+3指跳转到当前指令后面的第3条指令，以此类推
// 伪标号前与普通标号一样，仍需以冒号“:”开始
// 可以用“-1”代替当前指令的上一条指令，“-5”表示当前指令上面的第5条指令等
// 注意，这里的指令都是指有效指令，注释、标号等将被忽略（即不被算入）
// This example shows how to use "+1", "+3" and other "virtual-labels" or "pseudo-labels" in if and goto instructions to jump
// +1 refers to the next instruction that jumps to the current instruction,+3 refers to the third instruction after the current instruction, and so on
// The pseudolabel is the same as the ordinary label, and still needs to start with a colon ":"
// You can use "-1" instead of the last instruction of the current instruction, and "-5" means the fifth instruction above the current instruction, etc
// Note that the instructions here refer to valid instructions, and comments, labels, etc. will be ignored (that is, not counted)

// 将变量a赋值为字符串abc
// Assign variable a to string abc
assign $a "abc"

// 获取该字符串的长度，结果放入变量lenT中
// Get the length of the string and put the result into the variable lenT
len $lenT $a

// 判断lenT是否小于5，结果放入变量rsb中
// Judge whether lenT is less than 5, and put the result into the variable rsb
< $rsb $lenT #i5

// 如果rsb值为布尔值true，则跳转到下一条指令执行
// 否则跳转到下面第三条指令执行
// If the rsb value is a boolean value of true, skip to the next instruction execution
// Otherwise, skip to the third instruction below
if $rsb :+1 :+3
    pln "<5"

    // 无条件跳转到下面第二条指令
    // Unconditionally jump to the second instruction below
    goto :+2

    pln ">5"

pln a = $a
```

可以看出，直接使用伪标号跳转写法更简洁。不过也有不便之处，例如else分支如果用标号可能更方便，因为if分支如果要增减语句的话，else分支用数字就需要经常变化，容易遗漏出错。因此，可以结合普通标号与伪标号来使用跳转。

It can be seen that the jump writing method using pseudolabels directly is more concise. However, there are also inconveniences. For example, if the else branch is labeled, it may be more convenient, because if the if branch is to add or subtract statements, the numbers used for the else branch need to change frequently, which is easy to miss errors. Therefore, jump can be used in combination with common label and pseudolabel.

&nbsp;

##### - **for循环**（The for loop）

&nbsp;

有了条件分支结构，就可以实现标准的for循环，这是一个可以具有终止条件的循环语法结构。

With the conditional branch structure, you can implement the standard for loop, which is a loop syntax structure that can have termination conditions.

  ```go
// 实现类似 for i = 0; i < 5; i ++ 的标准三段for循环结构
// the following code has the same effect as the for-3 loop like: for i = 0; i < 5; i ++ {...}

// 将变量i赋值为整数0
// assign integer value 0 to variable i
assign $i #i0

// 标号loop1
// label loop1
:loop1

    // 将i的值加上整数10
    // 结果存入tmp
    // add 10 to $i
    // the result will be put into $tmp
    add $i #i10

    // 输出变量i中数值，和tmp值
    // output the result and value in $i
    pln $i ":"  $tmp

    // 将变量i的值加1
    // same as ++
    inc $i

    // 判断变量i中的数值是否小于整数5
    // if $i < 5, set $tmp to true
    < $i #i5

    // 是则跳转到标号loop1（继续循环）
    // if the value in $tmp is true, jump to label loop1
    if $tmp :loop1

// 否则执行下面的语句
// 也就是跳出了loop1的循环结构
// 输出字符串“end”
// else the code will coninue to run to the following line
pln end

  ```

上面的例子代码（for.xie）实现了一个经典的三段for循环结构。其中用到了inc指令，作用是将变量值加1，如果不带参数则会将\$tmp中的值加1，结果都将存入\$tmp。inc指令实现了一般语言中 i++ 的效果。本段代码执行的结果是：

The above example code (for.xie) implements a classic three-segment for loop structure. The inc instruction is used to add 1 to the variable value. If there is no parameter, the value in \$tmp will be added by 1, and the results will be stored in \$tmp. The inc instruction implements the effect of "i++" in other languages. The result of this code execution is:

  ```
    0 : 10
    1 : 11
    2 : 12
    3 : 13
    4 : 14
    end
  ```

与inc指令对应的是dec指令，会将对应值减1。

The inc instruction corresponds to the dec instruction, which will reduce the corresponding value by 1.

&nbsp;

##### - **利用for指令进行for循环**（Use the for instruction to carry out a for loop）

&nbsp;

谢语言也提供了for指令来进行常规的for循环，结合表达式可以实现灵活的循环控制，参看下面的例子（for3.xie）：

Xielang also provides for instructions to carry out regular 'for' loops. Combining with expressions can realize flexible loop control. See the following example (for3.xie):

```go
// 第一个循环开始
// 将变量i赋值为整数0
// the first loop starts here
// Assign variable i to integer 0
assign $i #i0

// 赋值用于循环终止条件判断的变量cond
// 赋值为布尔值true，以便第一次循环条件判断为true从而开始循环
// 否则一次都不会执行
// variable $cond is used for loop condition
// here we give it an initial value true(bool type)
// otherwise the first loop will not even run once
assign $cond #btrue

// 循环执行标号label1处的代码（即循环体）
// 直至变量cond的值为布尔值false
// 循环体中应该用continue指令继续循环或break中断循环
// 这是常见的三段式for循环，省略了第一和第三部分的代码（即不执行初始代码和每次循环后代码），等价于下面的代码：
// for loop
// Loop execution code at label1 (i.e. loop body)
// Until the value of variable cond is boolean false
// The continue instruction should be used in the loop body to continue the loop or break the loop
// This is a common three-segment 'for' loop, omitting the code in the first and third parts (that is, not executing the initial code and the code after each loop), which is equivalent to the following code:
//
// for (;cond;) { ... }
//
// 或者（or）:
// for cond { ... }
// 
for "" $cond "" :label1

// 第二个循环开始
// the second loop starts here

// 循环执行label2处代码
// 表达式是判断变量j小于2.8则执行label2处代码
// 这次将初始化循环变量的指令和给循环变量增长的指令放入了for指令中，构成了标准的三段式for循环结果
// 这条语句等价于C/C++语言中的：
// the quick eval expression in the first parameter(determines if variable $j < float value 2.8)
// the same as in C/C++：
//
// for (float j = 0.0; j < 2.8; j = j + 0.5) {...}
// 
// for指令最后两个标号分别是循环编号（循环条件满足时执行哪里）和跳出循环编号（即循环条件不满足时跳转到哪里），跳出循环编号可以省略，默认为:+1，即下一条语句
// The last two labels of the for instruction are the loop number (where to execute when the loop condition is satisfied) and the jump loop number (where to jump when the loop condition is not satisfied). The jump loop number can be omitted. The default is:+1, which is the next statement
for "assign $j #f0.0" @`$j < #f2.8` "add $j $j #f0.5" :label2 :+1

// （两个）循环结束输出
// the end of the both of the loops
pln "for end"

// 终止程序运行，否则将继续往下执行
// terminate the program, otherwise will run down to the following code
exit

// 第一个循环体代码
:label1
     // 输出变量i的值作参考
     pl "i=%v" $i

    // 将变量i的值加1
    // same as "++ $i" and "i++" in C/C++
    inc $i

    // 判断变量i的值是否小于整数5
    // 结果放入变量cond
    // check if $i < 5(int value) and put the bool result into $cond
    < $cond $i #i5

    // 继续执行循环（会再次判断for指令中的条件，结果为true才继续，否则跳出循环继续执行for指令后面的指令）
    // continue the loop(will check the condition defined in $cond again to determine if continue or end the loop and run the following code）
    // if break, default label is ":+1", means running the code line right after the "for" instruction
    continue

// 第二个循环体代码
// Second loop body code
:label2
    // 输出变量j的值作参考
    // The value of output variable j is used as reference
    pl j=%v $j

    // 继续执行循环（会再次判断for指令中的条件）
    // Continue to execute the loop (the condition in the for instruction will be judged again)
    continue

```

可以看出，谢语言中for指令做循环标准的写法如下：

It can be seen that the for instruction in Xielang is written as a loop standard as follows:

```
for "初始化指令" 判断条件 "循环间指令" 循环体标号 跳出循环标号
```

```
For "initialization instruction" condition "inter loop instruction" loopLabel breakLabel
```


for指令后面的第一个参数是循环前的初始化指令，第二个参数是循环条件，可以是一个表达式，满足条件（即值为布尔值true）才会进行循环；第三个参数是循环间指令，即每次循环体代码执行完后要执行的指令；第四个参数是一个标号，表示循环体代码开始的位置。循环体代码中应该用continue指令继续循环或break中断循环；第五个参数是跳出循环体的标号，即循环结束或跳出时要执行的代码，默认为:+1，即下一条语句。代码中演示了两个for循环，第一个for循环的循环条件是放在一个变量中的，第二个则直接用一个表达式来表示，每次循环都会重新计算这个表达式。运行结果如下：

The first parameter after the for instruction is the initialization instruction before the loop, and the second parameter is the loop condition, which can be an expression. Only when the condition is satisfied (that is, the value is a Boolean value of true) can the loop be performed; The third parameter is the inter loop instruction, which is the instruction to be executed after each loop body code execution; The fourth parameter is a label indicating the position where the loop body code starts. The loop body code should use the continue instruction to continue the loop or break to interrupt the loop; The fifth parameter is the label of the loop body that jumps out, which is the code to execute when the loop ends or jumps out. The default value is :+1, which indicates the next statement. The code demonstrates two for loops. The first for loop's loop conditions are placed in a variable, and the second is directly represented by an expression, which is recalculated each time the loop is executed. The operation results are as follows:

```shell
i=0
i=1
i=2
i=3
i=4
j=0
j=0.5
j=1
j=1.5
j=2
j=2.5
for end
```

&nbsp;

##### - **用range指令进行简单数据的遍历**（Iterating data using the range instruction）

&nbsp;

对于整数、字符串和一些简单的数组（后面会详细说明），可以用range指令对其进行遍历，即循环执行一定次数，每次循环体内可以获得遍历序号和遍历值进行相应操作，参看下面的代码（range.xie）：

For integers, strings, and some simple arrays (described in detail later), you can use the range instruction to traverse them, that is, execute the loop for a certain number of times, and each time the loop body can obtain the traversal number and traversal value for corresponding operations. See the following code (range.xie):

```go
// 循环遍历整数5，每次执行标号label1处的循环体代码
// 将循环5次，遍历值分别是0，1，2，3，4
// 相当于其他语言中的 for i := 1, i < 5; i ++……
// range/iterate integer 5, run the loop body at the position of label ":label1"
// then (break) run the code in label ":+1", ie. the next line of the "range" instr
// loop for 5 times, the iterated value will be 0，1，2，3，4
// the same as in C/C++: for i := 1, i < 5; i ++……
range #i5 :label1 :+1

// 第一个循环结束
// end of the first range
pln "end range1"

// 跳转到标号next1处
goto :next1

:label1
    // 用getIter指令获取遍历序号和遍历值
    // get the i, v from iterator
    // if iteration failed, $i will be an error object
    getIter $i $v
    checkErrX $i

    // 输出供参考
    pl "i: %v, v: %v" $i $v

    // 继续循环遍历
    // continue the loop/range
    continue

:next1

// 进行第二个循环，循环体在标号:label2处
// 第二个表示跳出循环的标号可以省略，默认为“:+1”
// 遍历字符串每次的遍历值是一个byte类型的整数
// the break label could be omitted, default is ":+1"
range "abc123" :label2

// 第二个循环结束
// end of the second range
pln "end range2"

// exit the program, or the next line will be run
exit

:label2
    // 用getIter指令获取遍历序号和遍历值
    getIter $i $v

    checkErrX $i

    pl "i: %v, v: %v" $i $v

    continue

```

执行结果是：

The execution result is:

```shell
i: 0, v: 0
i: 1, v: 1 
i: 2, v: 2 
i: 3, v: 3 
i: 4, v: 4 
end range1 
i: 0, v: 97
i: 1, v: 98
i: 2, v: 99
i: 3, v: 49
i: 4, v: 50
i: 5, v: 51
end range2 
```

注意遍历整数和字符串的区别。

Note the difference between traversing integers and strings.

最新版的谢语言中，range指令也支持切片（数组）和映射（字典）的遍历。

In the latest version of Xielang, the range instruction also supports traversal of slices (arrays) and maps (dictionaries).

&nbsp;

##### - **range嵌套**（range in range）

&nbsp;

另外请注意，range指令可以嵌套，如下所示（rangeInRange.xie）；

Note also that range instructions can be nested, as shown in the following(rangeInRange.xie);

```go
= $n1 #i5

range $n1 :range1

pln "end"

exit

:range1
    getIter $i $v

    pl "[1] %v: %v" $i $v

    range "abc" :range2

    continue


:range2
    getIter $j $jv

    pl "[2] %v: %v" $j $jv

    range @`$j + $j + #i1` :range3 :+1  $j @`$j + $j + #i1`

    continue

:range3
    getIter $k $kv

    pl "[3] %v: %v" $k @`$kv * $kv`

    continue

```

运行效果如下：

The operation effect is as follows:

```shell
[1] 0: 0
[2] 0: 97
[3] 0: 0 
[2] 1: 98
[3] 0: 1 
[3] 1: 4 
[2] 2: 99
[3] 0: 4 
[3] 1: 9 
[3] 2: 16
[1] 1: 1 
[2] 0: 97
[3] 0: 0 
[2] 1: 98
[3] 0: 1 
[3] 1: 4 
[2] 2: 99
[3] 0: 4 
[3] 1: 9 
[3] 2: 16
[1] 2: 2 
[2] 0: 97
[3] 0: 0
[2] 1: 98
[3] 0: 1
[3] 1: 4
[2] 2: 99
[3] 0: 4
[3] 1: 9
[3] 2: 16
[1] 3: 3
[2] 0: 97
[3] 0: 0
[2] 1: 98
[3] 0: 1
[3] 1: 4
[2] 2: 99
[3] 0: 4
[3] 1: 9
[3] 2: 16
[1] 4: 4
[2] 0: 97
[3] 0: 0
[2] 1: 98
[3] 0: 1
[3] 1: 4
[2] 2: 99
[3] 0: 4
[3] 1: 9
[3] 2: 16
end
```

&nbsp;

##### - **更多range数字的例子**（range numbers）

&nbsp;

下面是更多一些遍历数字的例子（rangeNumber.xie）；

```go
// 本例展示对整数或小数使用range指令遍历
// shows the range/iterate action of integer and float

// range指令后带一个整数则表示遍历整数5，循环体代码在标号label1处
// 跳出循环的标号在第三个，但可以省略，默认为“:+1”，即跳转到下一条指令继续执行
// 将依次输出每个循环序号和遍历值
// 遍历整数5相当于依次取0, 1, 2, 3, 4共5个遍历值
// 对应循环序号是0, 1, 2, 3, 4
// range integer 5
// range value for each round will be: 0, 1, 2, 3, 4
// range index will be: 0, 1, 2, 3, 4
range #i5 :label1

pln

// range指令后带两个整数表示范围遍历，此时跳出循环的标号不可省略
// 然后跟随遍历范围中的起始值，range后的数是结束值（不含）
// 这里是遍历整数2到5，也就是
// 依次取2, 3, 4共3个遍历值
// 对应循环序号是0, 1, 2
// range from 2 to 5
// range value for each round will be: 2, 3, 4
// range index will be: 0, 1, 2
// here the break label ":+1"(or other label) could not be omitted
range #i5 :label1 :+1 #i2

pln

// range指令后带三个整数表示指定步长的范围遍历
// 这里是遍历整数20到50，步长为5，也就是
// 依次取20, 25, 30, 35, 40, 45共6个遍历值
// 对应循环序号是0, 1, 2, 3, 4, 5
// 因此，完整的range指令应该类似：range 需遍历的值 继续遍历循环标号 跳出遍历循环标号 起始值 结束值（不含） 步长
// 对于数字的遍历，“需遍历的值”应于“结束值”一致，对于数组等的遍历，需遍历的值为数组等对象，起始值、结束值为整数
// range from 20 to 50, step 5(if not set, the default step is always 1)
// range value for each round will be: 20, 25, 30, 35, 40, 45
// range index will be: 0, 1, 2, 3, 4, 5
range #i50 :label1 :+1 #i20 #i50 #i5

pln

// 浮点数的遍历与整数类似，注意如果不指定步长，默认步长为1.0
// range of float value is the same as integer, if the step is not set, the default step is 1.0
// range value for each round will be: 0.2, 0.7, 1.2
// range index will be: 0, 1, 2
range #f1.6 :label1 :+1 #f0.2 #f1.6 #f0.5

pln

// 本例中步长为负值
// 将遍历12, 9, 6, 3, 0这5个值
// 并输出60除以这些值的结果
// 遍历过程中遇到为0的遍历值时，因为除数为零属于错误，会检查出错误信息并继续执行
// the step could be a negative value
// the last parameter is direction: 0(default): >=, 1: <=, 2: >, 3: <, 4: ==, means the condition operator to terminate the range loop
// in most cases, for positive range(0, 1, 2...), it should be 0, for negative range(9, 8, 7...), it will be 1
// range value for each round will be: 12, 9, 6, 3, 0
// the output will be 60 / range value
// range index will be: 0, 1, 2, 3, 4
// when range to value of 0, will trigger the error handler
range #i-9 :label2 :+1 #i12 #i-9 #i-3 1

exit

:label1
    // 遍历与循环一样，用getIter指令获取序号值与遍历值
    // get the range index and value
    getIter $i $v

    pln $i -> $v

    continue

:label2

    getIter $i $v

    div $rs1 #i60 $v

    ifErrX $rs1 :+1 :+3
        pl "failed to cal (60 / %v): %v" $v $rs1
        continue     
    
    pl "%v -> 60 / %v = %v" $i $v $rs1

    continue

```

运行结果是：

The operation effect is as follows:

```shell
0 -> 0
1 -> 1
2 -> 2
3 -> 3
4 -> 4

0 -> 2
1 -> 3
2 -> 4

0 -> 20
1 -> 25
2 -> 30
3 -> 35
4 -> 40
5 -> 45

0 -> 0.2
1 -> 0.7
2 -> 1.2

0 -> 60 / 12 = 5
1 -> 60 / 9 = 6
2 -> 60 / 6 = 10
3 -> 60 / 3 = 20
failed to cal (60 / 0): failed: runtime error: integer divide by zero(60, 0)
5 -> 60 / -3 = -20
6 -> 60 / -6 = -10
```

&nbsp;

##### - **switch分支**（switch branches）

&nbsp;

如同其他语言中的switch语句，谢语言中也支持用switch指令实现多种条件分支的便捷跳转，请看下面的例子（switch.xie）：

Like switch statements in other languages, Xielang also supports the use of switch instructions to achieve convenient jumps of multiple conditional branches. Please see the following example (switch.xie):

```go
= $a "abc"

// switch指令后，第一个参数是要判断的变量或数值
// 后面是一个个的数值与标号对，符合某个数值的情况下，将跳转到对应的标号
// 最后可以有一个单独的标号，表示默认跳转，即不符合任何条件的情况下跳转到哪里，默认是:+1，即下一条指令
// After the switch command, the first parameter is the variable or value to determine
// The following is a pair of numerical values and labels. If a certain value is met, it will jump to the corresponding label
// Finally, there can be a separate label indicating the default jump, that is, where to jump if any conditions are not met. The default is:+1, that is, the next instruction
switch $a "123" :label1  "abc" :label2 :label3

:label1
    pln label1
    exit

:label2
    pln label2
    exit

:label3
    pln label3
    exit

```

本例应该跳转到label2。

This example should jump to label2.

&nbsp;

##### - **switchCond分支**（switchCond branches）

&nbsp;

谢语言中还有一种类似switch指令的写法是switchCond指令，也可以实现多种条件分支的便捷跳转，比switch更接近于其他语言中的if...else if...else...结构，请看下面的例子（switchCond.xie）：

In Xielang, another method similar to the switch instruction is the switchCond instruction, which can also achieve convenient jumps of multiple conditional branches, which is closer to the structure of "if ... else if ... else..." in other languages, please see the following example (switchCond.xie):

```go
= $a "abc"

// switchCond指令后，是一个个的表达式（也可以是标识表达式的字符串）与标号对
// 这些表达式将一个个被依次计算，如果遇到有结果的布尔值为true的，将跳转到其对应的标号
// 最后可以有一个单独的标号，表示默认跳转，即不符合任何条件的情况下跳转到哪里，默认是:+1，即下一条指令
// After the switchCond instruction, there are expression and label pairs one by one
// These expressions will be evaluated one by one, and if a Boolean value with a result is true, it will jump to its corresponding label
// Finally, there can be a separate label indicating the default jump, that is, where to jump if any conditions are not met. The default is:+1, that is, the next instruction
switchCond @`$a == "123"` :label1  `$a > "aaa"` :label2 :label3

:label1
    pln label1
    exit

:label2
    pln label2
    exit

:label3
    pln label3
    exit

```

本例也会跳转到label2。

This example should jump to label2 as well.

&nbsp;

##### - **函数调用**（Function call）

&nbsp;

谢语言中的函数调用分为快速函数调用、一般函数调用和封装函数调用等多种方式，各有不同的优缺点，需要分别熟悉以便在不同场景下选择合适的调用方式。先介绍一般函数调用，一般函数调用的标准结构如下（func.xie）：
Function calls in Xielang can be divided into multiple methods, such as fast function calls, general function calls, and encapsulated function calls. Each method has different advantages and disadvantages, and it is necessary to be familiar with each method to select the appropriate call method in different scenarios. First, introduce general function calls. The standard structure of general function calls is as follows (func.xie):

```go
// 本例展示一般函数调用的方法
// 通过堆栈传入并传出参数
// This example shows the method of general function calls
// Pass in and out parameters through the stack

// 将变量s赋值为一个多行字符串
// Assign the variable s to a multiline string
assign $s ` ab c123 天然
森林 `

// 输出变量s中的值作为为参考
// plv指令会用内部表达形式输出后面变量中的值
// 例如会将其中的换行符等转义
// Print the value in the variable s for reference
// The plv instruction outputs the values in the following variables in an internal representation
// For example, it will escape line breaks and other characters
plv $s

// 将变量s中的值压栈
// Push the value in variable s onto the stack
push $s

// 调用函数func1
// 即跳转到标号func1处
// 而ret命令将返回到call语句的下一行有效代码处
// call指令后第一个参数为函数返回值，此处我们不用，所以用$drop变量将其丢弃
// Call function func1
// Jump to the label func1
// And the ret command will return to the next valid line of code in the call statement
// The first parameter after the call instruction is the function return value, which we do not use here, so we use the $drop variable to discard it
call $drop :func1

// 弹栈到变量s中，以便获取函数中通过堆栈传出的参数
// Pop the stack into the variable s to obtain the parameters in the function that pass through the stack
pop $s

// 再次输出变量s中的值
// Output the value in the variable s again
plv $s

// 终止代码执行
// Terminate code execution
exit

// 标号func1
// 也是函数的入口
// 一般称作函数func1
// Label func1
// Is also the entry point of the function
// It is commonly referred to as the function func1
:func1
    // 弹栈到变量v中，以便获取通过堆栈传入的参数
    // Pop the stack into the variable v to obtain the parameters passed through the stack
    pop $v

    // 将变量v中字符串做trim操作
    // 即去掉首尾的空白字符
    // 结果压入栈中
    // Trim the string in variable v
    // That is, remove the first and last white space characters
    // Results pushed onto the stack
    trim $push $v

    // 函数返回
    // 从相应call指令的下一条指令开始继续执行
    // Function return
    // Continue execution from the next instruction of the corresponding call instruction
    ret

```

上面代码中，plv指令会输出后面值的内部形式，主要为了调试时便于看出其中值的类型。call标号加ret指令是谢语言实现函数的基本方法，call语句将保存当前程序所处的代码位置，然后调用指定标号处的代码，直至ret语句时将返回到call时代码位置的下一条指令继续执行。这就实现了一个基本函数调用的逻辑。

In the above code, the plv instruction will output the internal form of the following values, mainly to facilitate the identification of the type of values during debugging. The call label plus the ret instruction is the basic method for Xielang to implement functions. The call statement will save the code location of the current program, and then call the code at the specified label until the ret statement returns to the next instruction at the code location at the time of the call to continue execution. This implements a basic function call logic.
    
&nbsp;

如果要给函数传递参数，可以通过堆栈来进行。同样地，函数返回值也可以通过堆栈来传递。trim指令实际上是对后面的变量进行去字符串首尾空白的操作，然后通过预置全局变量\$push进行压栈操作。

If you want to pass parameters to a function, you can do so through the stack. Similarly, function return values can also be passed through the stack. The trim instruction actually performs a whitespace removal operation on the following variables, and then performs a stack pushing operation by presetting the global variable \$push.

函数调用的其他方法将在后面逐一介绍。

The other methods of function calls will be described one by one later.

&nbsp;

##### - **函数调用时传递参数**（passing/retieving parameters in function call）

&nbsp;

上例中的函数调用，使用堆栈作为参数传递的路径。实际上，一般函数调用的参数传递还可以使用局部变量\$inputL和\$outL，这也是大多数函数调用时参数传递的常用方式。看下面的例子（func1.xie）来了解这种方式：

```go
// 本例展示一般函数调用的方法
// 通过$inputL和$outL堆栈来传入并传出参数
// This example shows the method of general function calls
// Pass in and out parameters through the $inputL and $outL stacks

// 将变量s赋值为一个多行字符串
// Assign the variable s to a multiline string
assign $s ` ab c123 天然
森林 `

// 输出变量s中的值作为为参考
// Print the value in the variable s for reference
plv $s

// 调用函数func1
// call指令后第一个参数为函数返回值，函数标号后的参数为输入参数，可以为0个、1个或多个
// 这些参数将以数组的形式传入到函数内的局部变量$inputL中
// Call function func1
// The first parameter after the call instruction is the function return value, and the parameters after the function label are the input parameters, which can be 0, 1, or more
// These parameters will be passed as an array into the local variable $inputL within the function
call $rs :func1 $s

// 再次输出变量s中的值
// Output the value in the variable s again
plv $rs

// 调用函数func2
// Call the function func2
call $rs :func2 $s

// 再次输出变量s中的值
// Output the value in the variable s again
plv $rs

// 终止代码执行
// Terminate code execution
exit

// 函数func1
// Function func1
:func1
    // 获取$inputL中的第一项，及我们传入的参数s
    // Get the first item in $inputL and the parameters we passed in, s
    getArrayItem $strL $inputL 0

    // 将变量strL中的字符串做trim操作
    // 结果存入变量outL中，这是约定的函数返回值的变量
    // Trim the string in variable strL
    // The result is stored in the variable outL, which is the default variable returned by the function
    trim $outL $strL

    // 函数返回
    // Function return
    ret

// 函数func2
// Function func2
:func2
    // 获取$inputL中的第一项，及我们传入的参数s
    // Get the first item in $inputL and the parameters we passed in, s
    getArrayItem $strL $inputL 0

    // 将变量strL中的字符串做trim操作
    // 结果存入变量outL中，这是约定的函数返回值的变量
    // Trim the string in variable strL
    // The result is stored in the variable rsL
    trim $rsL $strL

    // 函数返回时，如果ret指令后面带有一个参数，将被自动存入$outL中，达到函数返回值的目的
    // When a function returns, if the ret instruction is followed by a parameter, it will be automatically stored in $outL to achieve the purpose of returning the value of the function
    ret $rsL

```

本例演示了通过\$inputL和\$outL来传递函数的入参和出参，也演示了用ret指令后跟随出参的方式代替\$outL的方法。

This example demonstrates passing in and out parameters of a function through \$inputL and \$outL. It also demonstrates the method of replacing \$outL with a ret instruction followed by an out parameter.

&nbsp;

##### - **全局变量和局部变量**（Global and local variables）

&nbsp;

一般函数中会具有自己的局部变量空间，在函数中定义的变量（使用var指令），只能在函数内部使用，函数返回后将不复存在。而对变量值取值使用的情况，函数会先从局部变量寻找，如果有则使用之，如果没有该名字的变量则会到上一级函数（如果有的话，因为函数可以层层嵌套）中寻找，直至寻找到全局变量为止仍未找到才会返回“未定义”。对变量进行赋值操作的情况（对变量），如果在进入函数前没有定义过，则也会层层向上寻找，如果全没有找到，则会在本函数的空间内创建一个新的局部变量。如果要在函数中创建全局变量，则需要使用global指令。global指令与var指令用法一致，唯一的区别就是global指令将声明一个全局变量。看下面的例子（local.xie）来了解全局变量和局部变量的使用：

Generally, a function will have its own local variable space. Variables defined in the function (using the var instruction) can only be used inside the function, and will no longer exist after the function returns. For the use of variable values, the function will first search for local variables, if any, use them. If there is no variable with that name, it will search for a higher level function (if any, because the function can be nested hierarchically). Until the global variable is found and still not found, it will return "undefined". When performing an assignment operation on a variable (for a variable), if it has not been defined before entering the function, it will also be searched up layer by layer. If none is found, a new local variable will be created in the space of the function. If you want to create global variables in a function, you need to use the global directive. The global directive is used in the same way as the var directive, with the only difference being that the global directive declares a global variable. Look at the following example (local.xie) to understand the use of global and local variables:

```go
// 给全局变量a和b赋值为浮点数
// assign float values to variabe $a and $b
assign $a #f1.6
assign $b #f2.8

// 调用函数func1
// call function from label 'func1'
// and get the return result in variabe $rs
// the callFunction should return result in local variable $outL, or use instruction "ret" with a paramter for it
call $rs :func1

// 输出函数返回值
pln "function result:" $rs

// 输出调用函数后a、b、c、d四个变量的值
// output all the 4 variables after calling function :func1
pln $a $b $c $d

// 退出程序执行
// terminate the program
exit

// 函数func1
// function 'func1'
:func1
    // 输出进入函数时a、b、c、d四个变量的值
    // output all the 4 variables for reference
    pln $a $b $c $d

    // 将变量a与0.9相加后将结果再放入变量a中
    // add $a and float value 0.9, put the result again to $a
    add $a $a #f0.9

    // 声明一个局部变量b（与全局变量b是两个变量）
    // define a local variable with the same name $b as the global one(but they are completely different variables)
    var $b

    // 给局部变量b赋值为整数9
    // assign an integer value 9 to local variable $b
    assign $b #i9

    // 将局部变量b中的值加1
    // increase the number of local $b by 1
    inc $b

    // 将变量c赋值为字符串
    // assing a string value 'abc' to variable $c, also a local variable since not declared in global context
    = $c `abc`

    // 声明一个全局变量d
    // explicitly declare a global variabe $d
    global $d

    // 给变量d赋值为布尔值true
    // assign a bool value 'true' to global variable $d
    = $d #btrue

    // 退出函数时输出a、b、c、d四个变量的值
    // output all the 4 variables for reference
    pln $a $b $c $d

    // 函数返回，并带一个返回值
    // return from the function call, with a result value "done"
    ret "done"
```

注意其中的“=”是assign指令的另一种简便写法，另外assign指令前如果没有用global或var指令生命变量，相当于先用var命令声明一个变量然后给其赋值。这段代码的运行结果是：

Note that "=" is another convenient way to write the assign instruction. In addition, if the global or var instruction is not used before the assign instruction, it is equivalent to declaring a variable with the var command and assigning a value to it. The result of running this code is:

  ```
    1.6 2.8 未定义 未定义
    2.5 10 abc true
    2.5 2.8 未定义 true
  ```

注意其中4个变量a、b、c、d的区别，可以看出：变量a是在主代码中定义的全局变量，在函数func1中对其进行了计算（将a与0.9相加后的结果又放入a中）后，最后出了函数体之后的输出仍然是计算后的值，说明函数中操作的是全局变量；变量b则是在函数中定义了一个同名的局部变量，因此在函数中虽然有所变化，但退出函数后其值会变回原来的值，其实是局部变量b已经被销毁，此时的b是全局变量b；变量c完全是函数内的局部变量，因此入函数前和出了函数后都是“未定义”；变量c则是在函数中用global指令新建的全局变量，因此退出函数后任然有效。

Note the differences between the four variables a, b, c, and d, and it can be seen that variable a is a global variable defined in the main code. After calculating it in function func1 (adding the result of a and 0.9 into a), the output after the function body is finally output is still the calculated value, indicating that the function operates on global variables; Variable b refers to a local variable with the same name defined in the function. Therefore, although there are changes in the function, its value will change back to the original value after exiting the function. In fact, the local variable b has been destroyed, and at this time, b is the global variable b; The variable c is completely a local variable within the function, so both before and after entering the function are "undefined"; The variable c is a new global variable created in the function using the global instruction, so it remains valid after exiting the function.

&nbsp;

##### - **快速函数**（Fast functions）

&nbsp;

快速函数与一般函数的区别是：快速函数不会有自己的独立变量空间。快速函数与主函数（指不属于任何函数的代码所处的环境）共享同一个变量空间，在其中定义和使用的变量都将是全局变量。使用快速函数的好处是，速度比一般函数更快，因为减少了分配函数局部空间的开销。对一些实现简单功能的函数来说，有时候这是很好的选择。与快速函数传递参数可以使用堆栈或变量。

The difference between fast functions and general functions is that fast functions do not have their own independent variable space. The fast function shares the same variable space as the main function (the environment in which code that does not belong to any function), and the variables defined and used in it will be global variables. The advantage of using fast functions is that they are faster than regular functions because they reduce the overhead of allocating function local space. Sometimes this is a good choice for functions that implement simple functions. Passing parameters with fast functions can use a stack or variable.

快速函数类似call与ret的配对指令，使用fastCall与fastRet两个指令来控制函数调用与返回。下面是例子（fastCall.xie）：

The fast function is similar to the pairing instruction of call and ret, using two instructions, fastCall and fastRet, to control function calls and returns. The following is an example (fastCall.xie):

```go
// 将两个整数压栈
// push 2 integer values
push #i108
push #i16

// 快速调用函数func1
// 而fastRet命令将返回到fastCall语句的下一行有效代码处
// fast call func1
fastCall :func1

// 输出弹栈值（为函数func1压栈的返回结果）
// output the value upmost of the stack
plv $pop

// 终止代码执行
// terminate the program
exit

// 函数func1
// 功能是将两个数相加
// function func1
// add 2 nubmers
:func1
    // 弹栈两个数值
    // pop 2 values from stack to add
    pop $v2
    pop $v1

    // 将两个数值相加后压栈
    // add 2 nubmers and push the result to the stack
    add $push $v1 $v2

    // 函数返回（也可以用ret指令）
    // 从相应fastCall指令的下一条指令开始继续执行
    // return, continue to the next command of the fastCall(ret is also valid instead of fastRet)
    fastRet

```

运行结果为：

The running result is:

```shell
124
```

&nbsp;

##### - **寄存器**（Registers）

&nbsp;

如同汇编语言一样，谢语言也提供“寄存器”供便捷的存放和使用数值。谢语言中提供至少30个寄存器，按照数字索引使用，每个寄存器中可以存放一个数值。一定意义上，寄存器可以看做使用索引（而非名称）引用的全局变量。在编写简单的脚本时，有时候使用寄存器比使用变量更方便，代码更简洁。这里先给出一个寄存器使用的例子（reg.xie）以便理解：

Like assembly language, Xielang also provides "registers" for convenient storage and use of numerical values. Xielang provides at least 30 registers, which can be used according to numerical indexes, and each register can store a numerical value. In a certain sense, registers can be seen as global variables referenced using indexes rather than names. When writing simple scripts, sometimes using registers is more convenient and the code is simpler than using variables. Here is an example of using registers (reg.xie) to understand:

```go
// 本例演示使用寄存器来计算10的阶乘
// This example demonstrates using registers to calculate the factorial of 10

// 将编号为0的寄存器中存入整数1
// Store integer 1 in register number 0
= $#0 #i1

// 将编号为1的寄存器中存入整数1
// Store integer 1 in register number 1
= $#1 #i1

// 开始循环，loop1是循环开始的标号
// Start loop, loop1 is the label for the beginning of the loop
:loop1
    // 将寄存器1中的数值加1
    // Add 1 to the value in register 1
    inc $#1

    // 输出寄存器0和寄存器1中的数值作为参考
    // Output the values in register 0 and register 1 as references
    pln $#0 $#1
    
    // 将寄存器0中的数值与寄存器1中的数值相乘，结果存入寄存器0
    // Multiply the value in register 0 by the value in register 1, and store the result in register 0
    * $#0 $#0 $#1

    // 如果寄存器1中的数值大于等于10，则跳出循环
    // If the value in register 1 is greater than or equal to 10, jump out the loop
    if @"$#1 >= #i10" :end

    // 继续循环
    // Continue the loop
    goto loop1

:end
    // 中止程序运行
    // Exit the program
    exit

```

从上面的代码中可以看出，寄存器通过“\$#”加数字来引用，使用方法与变量一致，“\$#0”表示编号为0的寄存器（简称寄存器0），“\$#1”表示编号为1的寄存器（简称寄存器1）。30个寄存器的话，最后一个寄存器编号为29，即“\$#29”。

From the above code, it can be seen that registers are referenced by adding numbers to "\$#", using the same method as variables. "\$#0" represents the register with index 0 (referred to as register 0), and "\$#1" represents the register with index 1 (referred to as register 1). If there are 30 registers, the last register index is 29, i.e. "\$#29".

本段代码的运行结果是：

The running result of the code is as below:

```shell
1 2
2 3
6 4
24 5
120 6
720 7
5040 8
40320 9
362880 10

```

顺利地计算出了10的阶乘。

Successfully calculated the factorial of 10.

谢语言中，除了全局寄存器之外，每个函数上下文（函数上下文的详细概念请参见下一节）中，也提供一组局部寄存器（30个）供使用，仅在函数内有效。全局寄存器使用\$#加数字来引用，而局部寄存器使用\$~加数字来引用，例子可参见localReg.xie。

In Xielang, in addition to global registers, each function context (see the next section for detailed concepts of function context) also provides a set of local registers (30) for use, which are only valid within the function. Global registers are referenced using \$# plus numbers, while local registers are referenced using \$~ plus numbers. For example, see localReg.xie.

&nbsp;

##### - **谢语言的基础设施**（Infrastructure provided by Xielang）

&nbsp;

基础设施是计算机语言提供给开发者使用的各种工具和便利措施，基础设施一定程度上也影响语言代码的结构。谢语言提供了一系列基础设施以保证语言能力并提高开发效率，下面先介绍谢语言基础设施的整体结构，然后按照从底层到高层的顺序介绍谢语言中提供的基础设施。

Infrastructure is a variety of tools and convenience measures provided by computer language for developers to use. Infrastructure also affects the structure of language code to some extent. Xielang provides a series of infrastructure to ensure language proficiency and improve development efficiency. Below, we will first introduce the overall structure of Xielang's infrastructure, and then introduce the infrastructure provided in Xielang from the bottom to the top in order.

谢语言基础设施的整体结构包括：

The overall structure of Xielang Infrastructure includes:

- 跨虚拟机的全局上下文（Global Context Across Virtual Machines）
- 虚拟机（Virtual Machines）
- 运行上下文（Running Context）
- 函数上下文（Function Context）

一般来说，谢语言中代码是在单独的虚拟机中运行的，虚拟机中提供了堆栈和寄存器等供虚拟机内部的代码使用。一个虚拟机中包含默认的运行上下文和默认的“根”函数上下文。运行上下文包含了代码运行所需的基本信息，有时候也称为“运行环境”。函数是谢语言中组织代码的基本方式，函数上下文包含了函数中代码运行所需的变量等内容。谢语言中，直接编写的代码都是运行在默认运行上下文和根函数上下文中的，在某些情况下，代码也可以运行在不同的运行上下文中；而在进行函数调用操作时，一般代码将在不同函数上下文中运行。

Generally speaking, the code in Xielang runs in a separate virtual machine, which provides stacks and registers for internal code use. A virtual machine contains default runtime context and default 'root' function context. The runtime context contains the basic information required for code execution, sometimes also known as the "runtime environment". Functions are the basic way of organizing code in Xielang, and the function context includes variables and other content required for the code to run in the function. In Xielang, directly written code runs in the default runtime context and root function context, and in some cases, code can also run in different runtime contexts; When performing function call operations, the general code will run in different function contexts.

在并发调用或者某些特殊需要时，谢语言中也可以启动另一个虚拟机来执行代码。谢语言也提供一些跨虚拟机共享的基础设施供所有虚拟机中的代码使用，例如线程安全的队列、堆栈、序号发生器等。

In case of concurrent calls or certain special needs, Xielang can also start another virtual machine to execute code. Xielang also provides some shared infrastructure across virtual machines for code usage in all virtual machines, such as thread safe queues, stacks, sequence generators, etc.

下面详细介绍谢语言中的各个基础设施：

Below is a detailed introduction to the various infrastructure in Xielang:

- **函数上下文**：函数是谢语言中组织代码的基本方式，函数上下文即提供给某个函数中的代码使用的基础设施。函数上下文中的基础设施主要包括：局部变量（仅能再该函数内部使用，并可以被该函数中嵌套调用的函数使用）、\$tmp预置变量（、$tmp虽然说是预置全局变量，但为了尽量减少干扰，实际上是每个函数中有一个）、延迟调用栈（存放函数正常返回或异常退出时，会依次执行的一系列指令，参见defer指令相关的章节）。注意，即使没有显式地声明任何函数，所有代码也是在虚拟机默认提供的一个“根函数”的上下文中运行的，其他函数可以视为在根函数中嵌套调用的。

- **Function Context**: Functions are the basic way of organizing code in Xielang, and function context is the infrastructure provided for the use of code in a certain function. The infrastructure in the context of a function mainly includes: local variables (which can only be used internally within the function and can be used by nested functions within the function), \$tmp preset variables (although \$tmp is said to be a preset global variable, in order to minimize interference, there is actually one in each function) Delay the call stack (it stores a series of instructions that will be executed in sequence when the function returns normally or exits abnormally. See the chapter related to defer instructions). Note that even if no function is explicitly declared, all code runs in the context of a "root function" provided by the virtual machine by default, and other functions can be considered nested calls within the root function.

- **运行上下文**：运行上下文是谢语言中代码运行的必须的环境（因此有时候也称为运行环境），包含源代码、编译后的指令集、标号表、指令运行指针、函数栈、循环遍历栈等重要的运行时基础设施。虚拟机中提供一个默认的运行上下文，使用runCall等指令可以让代码在另一个运行上下文中运行。不同运行上下文有不同函数栈，除了虚拟机的根函数外，两个运行上下文中的函数互不干扰，但可以通过根函数中的变量等进行共享；因此，运行上下文可以实现将代码有效隔离起来。一个运行上下文中只有一个指令运行指针，意味着其中代码是无法并发执行的，因此如果要进行并发调用时，也可以使用goRunCall指令来创建新的运行上下文来实现，此时新的运行上下文将在不同线程中运行，但仍然可以与其他运行上下文共享根函数和虚拟机中的基础设施。

- **Running Context**: Running context is a necessary environment for code execution in Xielang(therefore, sometimes also known as the running environment), including important runtime infrastructure such as source code, compiled instruction sets, label tables, instruction run pointers, function stacks, and loop traversal stacks. Provide a default runtime context in the virtual machine, and use instructions such as runCall to allow code to run in another runtime context. Different running contexts have different function stacks. Except for the root function of the virtual machine, the functions in the two running contexts do not interfere with each other, but can be shared through variables in the root function; Therefore, the runtime context can effectively isolate the code. There is only one instruction running pointer in a running context, which means that the code cannot be executed concurrently. Therefore, if concurrent calls are to be made, the goRunCall instruction can also be used to create a new running context. At this time, the new running context will run in different threads, but it can still share the root function and infrastructure in the virtual machine with other running contexts.

- **虚拟机**：虚拟机是谢语言最基本的代码运行环境，虚拟机中除了前面提到的，提供一个默认的根函数上下文，以及一个默认运行上下文之外，还提供一个虚拟机内共享的堆栈，和30个寄存器（索引序号分别为0到29）。

- **Virtual Machine**: Virtual machine is the most basic code runtime environment for Xielang. In addition to providing a default root function context and a default runtime context, virtual machine also provides a shared stack within the virtual machine and 30 registers (index numbers 0 to 29, respectively).

- **跨虚拟机共享设施**：谢语言还提供一些跨虚拟机的全局共享设施，跨虚拟机的全局设施我们称之为“泛全局”设施，大多数情况下泛全局设施是为了并发处理时共享数据或者互通消息使用。这些共享设施都是并发（线程）安全的，包括一个队列、一个映射、一个堆栈、一个自增长序列号发生器、可定义的变量等。注意，泛全局设施仅在一个宿主进程中有效，例如两个谢语言主程序各自启动的虚拟机之间是无法共享泛全局设施的；但是一个谢语言主程序启动的多个虚拟机、某个虚拟机启动的嵌套虚拟机之间均可以使用泛全局共享设施。

- **Cross Virtual Machine Sharing Facilities**: Xielang also provides some global sharing facilities across virtual machines, which we call "pan global" facilities. In most cases, pan global facilities are used for sharing data or exchanging messages during concurrent processing. These shared facilities are all concurrency (thread) safe, including a queue, a map, a stack, a self growing sequence number generator, definable variables, and so on. Note that pan global facilities are only valid in one host process, for example, virtual machines launched by two Xielang main programs cannot share pan global facilities; However, multiple virtual machines initiated by a Xielang main program, as well as nested virtual machines initiated by a certain virtual machine, can all use pan global shared facilities.

Here we summarize the relationship between infrastructure in Xielang in a simple, less rigorous, but relatively easy to understand way: virtual machines contain a default runtime context and a function context (root function); A virtual machine can also contain more running contexts, each of which can contain multiple functions that are nested and called. The highest level of these functions in each running context is the only root function of the virtual machine; If we use function stacks to understand, each runtime context has a function stack, and the bottom layer of the function stack is the root function of the virtual machine. The meaning of running context is to isolate and run a piece of code, such as in the case of concurrent function calls.

&nbsp;

##### - **用runCall指令在不同运行上下文中执行代码**（Executing code in different running contexts using the runCall instruction）

&nbsp;

上面介绍了谢语言中的基础设施，了解了一个虚拟机中可以有多个运行上下文（简称运行环境）。常见的运行上下文的使用方式之一是使用runCall指令来调用函数，这种函数调用方式我们姑且称为runCall调用，下面是一个runCall调用的例子（runCall.xie）：

The above introduced the infrastructure in Xielang and learned that a virtual machine can have multiple runtime contexts (referred to as runtime environments). One of the common ways to use runtime context is to use the runCall instruction to call a function, which we can refer to as a runCall call. Here is an example of a runCall call (runCall.xie):

```go
// 编译一个函数
// 然后调用它，传入参数，接收返回值并处理
// runCall 指令将另起一个运行环境（上下文）与当前代码运行环境分离以避免冲突
// 两者仅共享虚拟机的基础设施（堆栈、寄存器）和全局变量
// 传递入参在函数中使用inputL变量引用，出参则通过outL变量赋值，如没有赋值outL，则返回undefined
// Compile a function
// Then call it, pass in parameters, receive the return value and process
// The runCall instruction separates another runtime environment (context) from the current code runtime environment to avoid conflicts
// Both only share the infrastructure (stack, registers) and global variables of the virtual machine
// Pass in the input parameter and use the inputL variable reference in the function, assign the output parameter to the outL variable, or return undefined if no outL value is assigned

compile $func1 `
    // 从$inputL中获取外界传入的两个参数
    // Obtain two parameters passed in from the outside world from $inputL
    getArrayItem $f1 $inputL 0
    getArrayItem $f2 $inputL 1

    pln arg1= $f1
    pln arg2= $f2

    // 将两个参数相加获取结果
    // Add two parameters to obtain the result
    add $result $f1 $f2

    // 在变量$outL中放入返回参数
    // Placing return parameters in variable $outL
    var $outL
    assign $outL $result

    exit
`

// 调用函数，并传入两个浮点数作为参数，返回结果存入$rs中
// Call the function, pass in two floating point numbers as parameters, and store the returned results in $rs
runCall $rs $func1 #f1.6 #f2.3

pln "runCall result:" $rs

```

这段代码的执行结果是：

The running result is:

```shell
arg1= 1.6
arg2= 2.3
runCall result: 3.9
```

上面是将一段文本代码编译后进行runCall调用。实际上，为了书写简单，我们也可以直接从代码中抽取一段形成一个运行上下文对象后进行runCall调用，参看下面的例子（runCall2.xie）：

The above is to compile a section of text code and make a runCall call. In fact, to simplify writing, we can also directly extract a segment from the code to form a runtime context object and make a runCall call, as shown in the following example (runCall2.xie):

```go

// extractRun指令将把从标号:func1Start开始到:func1End为止的指令转换为一个运行上下文以便调用
// The extractRun instruction will convert instructions from label :func1Start to :func1End into a running context for calling
extractRun $func1 :func1Start :func1End

// runCall指令可以直接将一个运行上下文当做一个函数来调用
// The runCall instruction can directly call a running context as a function
runCall $rs $func1 #f1.6 #f2.3

pln "runCall result:" $rs

// 注意要终止代码运行，否则将继续往下执行
// Be sure to terminate the code run, otherwise it will continue to execute
exit

// 标记函数开始的标号
// The label that marks the beginning of a function
:func1Start
    getArrayItem $f1 $inputL 0
    getArrayItem $f2 $inputL 1

    pln arg1= $f1
    pln arg2= $f2

    add $result $f1 $f2

    var $outL
    assign $outL $result

// 标记函数结束的标号
// The label that marks the end of the function
:func1End
    exit

```

这段代码的执行结果与前面一段是一样的。

The execution result of this code is the same as the previous one.

如上所示，runCall指令可以运行一段编译后的代码，也可以运行一个运行上下文。实际上runCall指令还可以直接跟一个起始标号和一个结束标号来调用一段函数代码，也可以直接调用一个字符串表示的函数代码。

As shown above, the runCall instruction can run a compiled piece of code or a runtime context. In fact, the runCall instruction can also directly call a section of function code with a starting label and an ending label, or directly call a function code represented by a string.

注意，runCall指令调用的函数，使用\$inputL局部变量来传入参数，这是一个数组（列表），需要用getArrayItem指令按索引从其中获取各个参数，使用\$outL参数来传出返回参数。另外，runCall调用的函数中，是可以访问所在虚拟机中定义的全局变量的。

Note that the function called by the runCall instruction uses the \$inputL local variable to pass in parameters, which is an array (list). The getArrayItem instruction needs to retrieve each parameter by index from it, and the \$outL parameter needs to be used to pass out return parameters. In addition, the functions called by runCall can access global variables defined in the virtual machine where they are located.

由于runCall指令调用的函数是在同一个虚拟机中运行的，因此可以使用本虚拟机的堆栈，下面是再一个的runCall的例子：

Since the function called by the runCall instruction runs in the same virtual machine, the stack of this virtual machine can be used. Here is another example of runCall:

```go
// 本例继续演示用runCall指令调用函数的方法
// This example continues to demonstrate the method of calling functions using the runCall instruction

// 压栈准备传入的函数的第一个参数
// Push the stack to prepare the first parameter of the incoming function
push #f1.6

// runCall指令将代码块看做封装函数进行调用
// 结果参数不可省略，之后第1个参数表示函数代码
// 之后可以跟多个指令参数，表示传入这个函数内的参数
// 这些参数在函数代码内可以通过列表类型的变量inputL访问
// 下面的这个函数的功能是简单的加法运算
// The runCall instruction treats code blocks as encapsulated functions for calling
// The result parameter cannot be omitted, and the first parameter after it represents the function code
// Afterwards, multiple instruction parameters can be used to represent the parameters passed into this function
// These parameters can be accessed within the function code through the variable inputL of the list type
// The function below is for simple addition operations
runCall $rs `
    // 弹栈第一个参数
    // pop out the first parameter
    pop $arg1

    // 获取传入的参数作为加法计算的第二个参数
    // Obtain the passed in parameter as the second parameter for addition calculation
    getArrayItem $arg2 $inputL 0

    // 输出两个参数检查
    // Output two parameter checks
    pln arg1= $arg1
    pln arg2= $arg2

    // 将两个参数相加，结果压栈
    // Add two parameters and stack the result
    add $push $arg1 $arg2

    // 输出栈顶值检查
    // Output stack top value check
    pln $peek

    // runCall函数因为与调用者在同一个虚拟机中运行，所以返回值可以通过堆栈返回
    // 也可以通过outL变量来返回值
    // The runCall function runs in the same virtual machine as the caller, so the return value can be returned through the stack
    // You can also return values through the outL variable
    assign $outL $peek
` #f2.3

// 输出函数返回值
// Output function return value
pln "result:" $rs

// 输出从堆栈返回的值，这个例子里是一样的
//Output the value returned from the stack, which is the same in this example
pln "result:" $pop

```

运行结果是：

```shell
arg1= 1.6
arg2= 2.3
3.9
result: 3.9
result: 3.9
```

后面在合适的时候，我们还将介绍使用不同运行上下文的并发调用指令goRunCall。

Later, when appropriate, we will also introduce the use of the concurrent call instruction goRunCall with different running contexts.

&nbsp;

##### - **取变量引用及取引用对应变量的实际值**（Reference and Dereference）

&nbsp;

这里的“引用”可以理解成一般语言中的取变量地址的操作。使用引用的目的是为了直接修改其中的值，尤其是对一些复杂数据类型来说。这里先给出一个对基础数据类型的取引用与解引用操作的例子（refUnref.xie）：

The 'reference' here can be understood as the operation of taking variable addresses in general language. The purpose of using references is to directly modify the values, especially for some complex data types. Here is an example of reference and dereference operations for basic data types (refUnref.xie):

```go
// 新建一个整数类型的引用变量$ref1
// 此时该引用将指向一个已经分配好可以容纳一个整数空间的地址
// Create a new reference variable $ref1 of integer type 
// At this point, the reference will point to an address that has already been allocated to accommodate an integer space
new $ref1 int

// 将该引用指向的地址中存入整数9
// Store the integer 9 in the address pointed to by this reference
assignRef $ref1 #i9

// 将$ref1指向的整数值取出存入$v中
// 这一步也称作将$ref1解引用
// Extract the integer value pointed to by $ref1 and store it in $v
// This step is also known as dereferencing $ref1
unref $v $ref1

// 输出$v的值作为参考
// Output the value of $v as a reference
plo $v

// 变量前加*，表示将其解引用
// 输出的值应该与$v一样
// Adding a "*" before a variable indicates its dereference
// The output value should be the same as $v
plo *$ref1

// 调用函数func1，并将引用变量$ref1作为参数传入
// Call function func1 and pass in the reference variable $ref1 as a parameter
call $rs :func1 $ref1

// 输出调用函数func1后$ref1变量的引用值与解引用值
// Output the reference and dereference values of the $ref1 variable after calling function func1
plo $ref1
plo *$ref1

// 退出程序执行
// Exit the program
exit

// 函数func1
// Function func1
:func1

    // 获取第一个输入参数
    // 即传入的引用变量$ref1
    // Get the first input parameter
    // That is, the passed in reference variable $ref1
    getArrayItem $p $inputL 0

    // 输出变量p参考
    // Output $p for reference
    plo $p

    // 将引用变量p中的对应的数值放入变量v中
    // Place the corresponding numerical value in the reference variable p into the variable v
    unref $v $p

    // 输出变量v
    // Output $v
    plo $v

    // 尝试将引用变量p中存放的实际数值重新置为浮点数1.6
    // 由于$p是个整数的引用，因此实际上存入的将是截取了小数点后部分的整数1
    // Try to reset the actual value stored in the reference variable p to floating point 1.6
    // Since $p is a reference to an integer, what is actually stored will be the integer 1 that has been truncated after the decimal point
    assignRef $p #f1.6

    // 函数返回
    // return from the function
    ret

```

代码中有详细的注释，运行结果为：

There are detailed comments in the code, and the running result is:

```shell
(int)9
(int)9
(*int)(*int)(0xc0003bc940)
(int)9
(*int)(*int)(0xc0003bc940)
(int)1
```

注意，其中的地址值可能随每次运行有所变化。

Note that the address values may vary with each run.

其中，ref指令用于取变量的引用，unref指令用于获取引用变量指向的值（解引用），assignRef指令则直接将引用变量指向的值赋以新值。可以看出，使用变量引用，成功地在函数中将全局变量中的数值进行了改变。

Among them, the ref instruction is used to obtain the reference of a variable, the unref instruction is used to obtain the value pointed to by the reference variable (dereference), and the assignRef instruction directly assigns a new value to the value pointed to by the reference variable. It can be seen that using variable referencing successfully changed the numerical values in the global variable in the function.

另外，new \$ref1 int也可以用var \$ref1 "\*int"来达到同样的效果，还可以用var \$ref1 "\*int" \#i23来进行初始化赋值（注意赋的是引用指向的数值而非指针地址）。

In addition, new \$ref1 int can also be used with var \$ref1 "\*int" to achieve the same effect, and var \$ref1 "\*int" \#i23 can also be used for initialization assignment (note that the assignment is to the numerical value pointed to by the reference rather than the pointer address).

&nbsp;

##### - **取谢语言变量的引用及其解引用**（Reference variable in Xielang and Dereference it）

&nbsp;

上一节介绍的变量引用与解引用指的是当变量中保存的是一个引用（指针）变量时，如何取引用与解引用。如果变量中保存的是一个普通的数值，我们想对变量本身进行取引用以便在别处使用（例如在不同线程中共享），那么需要使用refVar指令，相应地还有unrefVar指令用于解引用变量，以及assignRefVar指令用于根据引用进行复制。我们来看下面这个例子（refUnref2.xie）：

The variable reference and dereference introduced in the previous section refer to how to take a reference and dereference when a reference (pointer) variable is stored in the variable. If a variable holds a regular value in Xielang and we want to take a reference to the variable itself for use elsewhere (such as sharing in different threads), then we need to use the refVar instruction, correspondingly, the unrefVar instruction for dereferencing the variable, and the assignRefVar instruction for copying based on the reference. Let's take a look at the following example (refUnref2.xie):

```go
// 将变量a赋值为一个映射类型的数据
// Assign variable a as data of a mapping type
= $a #M`{"key1":1,"key2":"2"}`

// 输出a中的值作为参考
// output value in a for reference
pl "[1] a=%#v" $a

// 获取变量a的引用，放在变量p1中
// Obtain a reference to variable a and place it in variable p1
refVar $p1 $a

pl "p1=%#v" $p1

// 调用p1的成员方法GetValue来获取其中保存的值放在v1中
// 此时v1应该与a中的值是一样的
// Call the member method GetValue of p1 to obtain the saved value and place it in v1
// At this point, v1 should be the same as the value in a
mt $v1 $p1 GetValue

pl "v1=%#v" $v1

// 将v1中键名为key1的键值置为整数9
// Set the key value with key1 in v1 to the integer 9
setMapItem $v1 key1 #i9

pl "*p1=%#v" *$p1
pl "[2] a=%#v" $a

// 调用引用变量p1的成员方法SetValue将其重新赋值为浮点数3.6
// Call the member method SetValue of reference variable p1 to reassign it to floating point 3.6
mt $rs1 $p1 SetValue #f3.6

pl "[3] a=%#v" $a

// 解引用p1，将其放于变量rs2中
// 此时rs2中的值就是浮点数3.6
// Dereference p1 and place it in variable rs2
// At this time, the value in rs2 is floating point 3.6
unrefVar $rs2 $p1 

pl "rs2=%v" $rs2

// 用assignRefVar指令给引用变量指向的变量赋值
// Assign a value to the variable pointed to by the reference variable using the assignRefVar instruction
assignRefVar $p1 #btrue

pl "[4] a=%#v" $a

```

运行结果如下：

The running result will be:

```shell
[1] a=map[string]interface {}{"key1":1, "key2":"2"}
p1=&tk.FlexRef{Type:"map", Data:map[string]interface {}{"a":map[string]interface {}{"key1":1, "key2":"2"}, "argsG":[]string{"xie", "-gopath", "refUnref2.xie"}, "backQuoteG":"`", "guiG":(tk.TXDelegate)(0xfec300), "newLineG":"\n", "nilG":interface {}(nil), "p1":(*tk.FlexRef)(0xc00009e5c0), "scriptPathG":"D:\\goprjs\\src\\github.com\\topxeq\\xie\\cmd\\scripts\\refUnref2.xie", "undefinedG":tk.UndefinedStruct{int:0}}, Key:"a", Index:0}
v1=map[string]interface {}{"key1":1, "key2":"2"}
*p1=map[string]interface {}{"key1":9, "key2":"2"}
[2] a=map[string]interface {}{"key1":9, "key2":"2"}
[3] a=3.6
rs2=3.6
[4] a=true

```

&nbsp;

##### - **复杂数据类型-列表**（Complex Data Types - List）

&nbsp;

列表在其他语言中有时候也称作“数组”、“切片”等。在谢语言中，列表可以理解为可变长的数组，其中可以存放任意类型的值。列表的操作包括创建、增加项、删除项、切片（截取其中一部分）、合并（与其他列表合并）、遍历（逐个对列表中所有的数据项进行操作）等。下面的代码演示了这些操作的方法（list.xie）：

Lists are sometimes referred to as "arrays", "slices", etc. in other languages. In Xielang, a list can be understood as a variable length array that can hold any type of value. The operations of a list include creating, adding items, deleting items, slicing (taking a part of it), merging (merging with other lists), traversing (operating on all data items in the list one by one), and so on. The following code demonstrates the methods of these operations (list.xie):

```go
// 定义一个列表变量list1
// Define a list variable list1
var $list1 list

// 查看列表对象，此时应为空的列表
// View the list object, which should be an empty list at this time
plo $list1

// 给列表list1中添加一项整数8
//Add an integer 8 to list1
addItem $list1 #i8

// 给列表list1中添加一项浮点数12.7
// Add a floating point number 12.7 to the list list1
addItem $list1 #f12.7

// 再次查看列表list1中内容，此时应有两项
// Review the contents of list 1 again, and there should be two items at this time
plo $list1

// 用赋值的方法直接将一个数组赋值给列表变量list2
// #号后带大写的L表示后接JSON格式表达的数组
// Directly assign an array to the list variable list2 using the assignment method
// The L with uppercase after the # sign represents an array followed by JSON format expression
assign $list2 #L`["abc", 2, 1.3, true]`

// 输出list2进行查看
// Output list2 for viewing
plo $list2

// 查看list2的长度（即其中元素的个数）
// View the length of list2 (i.e. the number of elements in it)
len $list2

pln length= $tmp

// 获取列表list1中序号为0的项（列表序号从零开始，即第1项）
// 结果将入栈
// Get the item with sequence number 0 in list 1 (the list sequence number starts from zero, which is the first item)
// The result will be stacked
getItem $push $list1 #i0

// 获取list2中的序号为1的项，结果放入变量a中
// Obtain the item with sequence number 1 in list2 and place the result in variable a
getItem $a $list2 #i1

// 将变量a转换为整数（原来是浮点数）并存回a中
// Convert variable a to an integer (originally a floating point number) and return it to a
convert $a $a int

// 查看变量a中的值
// View the value in variable a
plo $a

// 将弹栈值（此时栈顶值是列表list1中序号为0的项）与变量a相加
// 结果压栈
// Add the stack value (where the top value of the stack is the item with number 0 in list 1) to variable a
// Result Stacking
add $push $pop $a

// 查看弹栈值
// Check the stack top value
plo $pop

// 将列表list1与列表list2进行合并
// 结果放入新的列表变量list3中
// 注意，如果没有指定结果参数（省略第一个，此时应共有2个参数），将把结果存回list1
// 相当于把list1加上了list2中所有的项
// Merge list 1 with list 2
// Place the results in the new list variable list3
// Note that if no result parameter is specified (omitting the first one, there should be a total of 2 parameters), the result will be saved back to list1
// Equivalent to adding all the items in list1 and list2
addItems $list3 $list1 $list2

// 查看列表list3的内容
// View the contents of list 3
plo $list3

// 将list3进行切片，截取序号1（包含）至序号5（不包含）之间的项
// 形成一个新的列表，放入变量list4中
// Slice list3 and extract items between numbers 1 (inclusive) and 5 (exclusive)
// Form a new list and place it in variable list4
slice $list4 $list3 #i1 #i5

// 查看列表list4的内容
// View the contents of list 4
plo $list4

// 循环遍历列表list4中所有的项，对其调用标号range1开始的代码块
// 该代码块必须使用continue指令继续循环遍历
// 或者break指令跳出循环遍历
// 遍历完毕或者break跳出遍历后，代码将继续从rangeList指令的下一条指令继续执行
// 遍历每项时，range指令指向的遍历体代码应用getIter指令获得当前当前序号值（从0开始）和遍历项的值
// 并保证用continue进行继续遍历，遍历完毕后将默认返回range指令后的下一条指令继续执行
// Loop through all items in list 4 and call the code block starting with the label range1 on them
// This code block must continue to loop through using the continue instruction
// Or break instruction jumps out of loop traversal
// After the traversal is completed or the break jumps out, the code will continue to execute from the next instruction in the rangeList instruction
// When traversing each item, the range instruction points to the traversal body code and applies the getIter instruction to obtain the current ordinal value (starting from 0) and the value of the traversal item
// And ensure to continue the traversal using continue. After the traversal is completed, the next instruction after the range instruction will be returned by default to continue execution
range $list4 :range1

// 删除list4中序号为2的项(此时该项为整数2)
// Delete the item with sequence number 2 in list 4 (at this point, the item is an integer 2)
deleteItem $list4 #i2

// 再次删除list4中序号为2的项(此时该项为浮点数1.3)
// Delete the item with No. 2 in list4 again (at this time, the item is floating point 1.3)
deleteItem $list4 #i2

// 修改list4中序号为1的项为字符串“烙红尘”
// Modify the item with serial number 1 in list 4 to the string "烙红尘"
setItem $list4 #i1 烙红尘

// 再次删除list4中序号为0的项(此时该项为浮点数12.7)
// Delete the item with serial number 0 in list4 again (at this time, the item is a floating point 12.7)
deleteItem $list4 #i0

// 再次查看列表list4的内容
// 此时应只剩1项字符串“烙红尘”
// Review the contents of list 4 again
// At this point, there should only be one string left, "烙红尘"
plo $list4

// 结束程序的运行
// Exit the program
exit

// 标号range1的代码段，用于遍历列表list4
// Code snippet labeled range1 for traversing list list4
:range1
    // 获得遍历序号和遍历值分别放入变量i和v中
    // Obtain the traversal sequence number and traversal value and place them in variables i and v, respectively
    getIter $i $v

    // 判断i值是否小于3，结果存入$tmp
    // Determine if the value of i is less than 3, and store the result in $tmp
    < $i #i3

    // 如果是则跳转到next1（继续执行遍历代码）
    // If true, jump to label next1 (continue the iterating code)
    if $tmp :next1

        // 否则跳出循环遍历
        // Otherwise, skip the loop traversal
        break

    // 标号next1
    // Label next1
    :next1

    // 输出提示信息
    // output the info
    pl `第%v项是%v` $i $v

    // 继续循环遍历，如欲跳出循环遍历，可以使用break指令
    // Continue loop traversal, if you want to break out of the loop traversal, you can use the break instruction
    continue
```

代码中有详细注释，运行的结果是：

There are detailed comments in the code, and the result of running is:

```shell
  ([]interface {})[]
  ([]interface {})[8 12.7]
  ([]interface {})[abc 2 1.3 true]
  length= 4
  (int)2
  (int)10
  ([]interface {})[8 12.7 abc 2 1.3 true]
  ([]interface {})[12.7 abc 2 1.3]
  第0项是12.7
  第1项是abc
  第2项是2
  ([]interface {})[烙红尘]
```

谢语言还有其他类型的列表，包括字节列表（byteList）和如痕列表（runeList）等，用法类似。

Xielang also has other types of lists, including byteList and runeList, with similar usage.

&nbsp;

##### - **复杂数据类型-映射**（Complex Data Types - Map）

&nbsp;

映射在其他语言中也称作字典、哈希表等，其中存储的是一对对“键（key）”与“值（value）”，也称为键值对（key-value pair）。谢语言中运用映射各种基本操作的例子如下（map.xie）：

Map, also known as dictionary, hash table, etc. in other languages, stores a pair of "key" and "value" pairs, also known as key value pairs. An example of using map for various basic operations in Xielang is as follows (map.xie):

```go
// 定义一个映射变量map1
// Define a map variable map1
var $map1 map

// 查看映射对象，此时应为空的映射
// View the map object, which should be an empty map at this time
plo $map1

// 给映射map1中添加一个键值对 “"Name": "李白"”
// setItem也可用于修改
// Add a key value pair 'Name' to map1: '李白'
// SetItem can also be used to modify values in map
setMapItem $map1 Name "李白"

// 再给映射map1中添加一个键值对 “"Age": 23”
// 此处23为整数
// Add a key value pair "Age": 23" to map1
// Here 23 is an integer
setMapItem $map1 Age #i23

// 再次查看映射map1中内容，此时应有两个键值对
// Looking at the content in map1 again, there should be two key value pairs at this point
plo $map1

// 用赋值的方法直接将一个数组赋值给映射变量map2
// #号后带大写的M表示后接JSON格式表达的映射
// Directly assign an array to the variable map2 using the assignment method
// The M with uppercase after the # sign represents a map followed by JSON format expression
assign $map2 #M`{"日期": "2022年4月23日","气温": 23.3, "空气质量": "良"}`

// 输出map2进行查看
// Output map2 for viewing
plo $map2

// 查看map2的长度（即其中元素的个数）
// View the length of map2 (i.e. the number of elements in it)
len $map2

pln length= $tmp

// 获取映射map1中键名为“Name”的项
// 结果入栈
// Obtain the item with key name 'Name' in map1
// Then push the result value to the stack
getMapItem $push $map1 Name

// 获取map2中的键名为“空气质量”的项，结果放入变量a中
// Obtain the key named "空气质量" in map2 and place the result in variable a
getMapItem $a $map2 空气质量

// 将弹栈值（此时栈顶值是映射map1中键名为“Name”的项）与变量a相加
// 结果压栈
// Add the stack value (where the top value of the stack is the key named "Name" in map1) to the variable a
// Then push the result value to the stack
add $push $pop $a

// 查看弹栈值
// View the popped value in stack
plo $pop

// 循环遍映射map2中所有的项，对其调用标号range1开始的代码块
// 该代码块必须使用continue指令继续循环遍历
// 或者break指令跳出循环遍历
// 遍历完毕或者break跳出遍历后，代码将继续从rangeMap指令的下一条指令继续执行
// Loop through all the items in map2 and call the code block starting with the label range1 on it
// This code block must continue to loop through using the continue instruction
// Or break instruction jumps out of loop traversal
// After the traversal is completed or the break jumps out, the code will continue to execute from the next instruction of the rangeMap instruction
range $map2 :range1

// 删除map2中键名为“气温”的项(此时该项为浮点数23.3)
// Delete the item named "气温" in map2 (at this time, it is a floating point 23.3)
deleteMapItem $map2 "气温"

// 再次查看映射map2的内容
// View the content in map2 again
plo $map2

// 结束程序的运行
// Exit the program
exit

// 标号range1的代码段，用于遍历映射
// Label range, used to iterating the map
:range1
    // 用getIter指令获取遍历序号与遍历项的值
    // Using the getIter instruction to obtain the traversal sequence number and the value of the traversal term
    getIter $k $v

    // 输出提示信息
    // Output the information for reference
    pl `键名为 %v 项的键值是 %v` $k $v

    // 继续循环遍历，如欲跳出循环遍历，可以使用break指令
    // Continue loop traversal, if you want to break out of the loop traversal, you can use the break instruction
    continue

```

其中详细介绍了映射类型的主要操作，代码的运行结果是：

The main operations of mapping types are described in detail, and the running results of the code are:

```shell
(map[string]interface {})map[]
(map[string]interface {})map[Age:23 Name:李白]
(map[string]interface {})map[日期:2022年4月23日 气温:23.3 空气质量:良]
length= 3
(string)李白良
键名为 日期 项的键值是 2022年4月23日
键名为 气温 项的键值是 23.3
键名为 空气质量 项的键值是 良
(map[string]interface {})map[日期:2022年4月23日 空气质量:良]
```


&nbsp;

##### - **取列表项、映射项和列表切片的快捷写法**（Shortcut for taking list items, mapping items, and list slices）

&nbsp;

谢语言中，取列表项除了使用getArrayItem指令外，可以用下述写法来直接作为参数在指令中使用：

In Xielang, in addition to using the getArrayItem instruction, taking a list item can be directly used as a parameter in the instruction using the following notation:

```go
assign $list2 #L`["abc", 2, 1.3, true]`

plo [$list2,2]
```

将取到列表的第三项（序号为2），输出：

Take the third item in the list (with a sequence number of 2) and output:

```shell
(float64)1.3
```

相应地，取映射项可以写作：

Correspondingly, taking the map item can be written as:

```go
assign $map2 #M`{"Date": "2022.4.23","Temperature": 23.3, "Air": "Good"}`

pln {$map2,Air}

```

会输出Good。

Will output 'Good'.

列表切片的写法类似：

The way of list slicing is similar to:

```go
assign $list2 #L`["abc", 2, 1.3, true]`

plo [$list2,1,3]
```

会输出：

Will output:

```shell
([]interface {})[]interface {}{2, 1.3}
```

&nbsp;

##### - **嵌套的复杂数据结构及JSON编码**（Nested complex data structures and JSON encoding）

&nbsp;

谢语言中，复杂数据结构也是可以嵌套的，例如列表中的数据项可以是一个映射或列表，映射中的键值也可以是列表或映射。看下面的例子（toJson.xie）：

In Xielang, complex data structures can also be nested, for example, data items in a list can be a map or list, and key values in the map can also be a list or map. Take the following example (toJson.xie):

```go
var $map1 map

setMapItem $map1 "姓名" 张三

setMapItem $map1 "年龄" #i39

var $map2 map

setMapItem $map2 "姓名" 张胜利

setMapItem $map2 "年龄" #i5

var $list1 list

addItem $list1 $map2

setMapItem $map1 "子女" $list1

plo $map1

toJson $push $map1 -indent -sort

pln $pop

```

例子中创建了一个简单的父子关系的数据结构，父亲张三，孩子张胜利，父亲这个数据对象本身是用映射来表示的，而其子女是用列表来表示，列表中的数据项——他的孩子张胜利本身又是用一个映射来表示的。另外，为了展示更清楚，我们使用了toJson指令，这个指令可以将数据结构转换为JSON格式的字符串，第一个参数是结果放入的变量，这里用内置变量$push表示将结果压栈。目前，toJson函数支持两个可选参数，-indent表示将JSON字符串用缩进的方式表达，-sort表示将映射内的键值对按键名排序。代码运行结果如下：

In the example, a simple parent-child relationship data structure was created, with father 张三 and child 张胜利. The father data object itself is represented by a map, while its children are represented by a list. The data items in the list - his child 张胜利 itself - are represented by a map. In addition, for clarity, we used the toJson instruction, which can convert data structures into JSON formatted strings. The first parameter is the variable into which the result is placed. Here, the built-in variable \$push is used to push the result onto the stack. Currently, the toJson function supports two optional parameters: - indent represents the indentation of JSON strings, and - sort represents the sorting of key values within the mapping by key names. The code run results are as follows:

```shell
(map[string]interface {})map[姓名:张三 子女:[map[姓名:张胜利 年龄:5]] 年龄:39]
{
  "姓名": "张三",
  "子女": [
  {
    "姓名": "张胜利",
    "年龄": 5
  }
],
  "年龄": 39
}
```

注意对比谢语言对该数据的表达形式与JSON形式的区别。

Pay attention to the difference between Xielang's expression of this data and JSON format.

&nbsp;

##### - **JSON解码**（JSON decoding）

&nbsp;

我们对JSON编码的反操作就是将JSON格式的字符串转换为内部的数据。这可以通过定义参数时加上“#L”或“#M”形式来进行，也可以通过fromJson指令来执行。使用“#L”或“#M”的方式我们前面已经介绍过了，这里是使用fromJson指令的例子，我们就直接用上面生成的JSON来反向操作试一下（fromJson.xie）：

Our reverse operation of JSON encoding is to convert JSON formatted strings into internal data. This can be done by adding "#L" or "#M" forms when defining parameters, or by executing the from Json instruction. We have already introduced the method of using "#L" or "#M". Here is an example of using the "fromJson" instruction. We will directly use the JSON generated above to reverse the operation and try it out (fromJson.xie):

```go
// 将变量s赋值为一个多行字符串
// 即所需解码的JSON文本
// Assign variable s as a multiline string
// The JSON text that needs to be decoded
assign $s `
{
  "姓名": "张三",
  "子女": [
  {
    "姓名": "张胜利",
    "年龄": 5
  }
],
  "年龄": 39
}
`

// 用fromJson指令将s中的文本解码到变量map1中
// Decode the text in s into variable map1 using the fromJson instruction
fromJson $map1 $s

// 获取map1的数据类型
// 可用于以后根据不同类型进行不同处理
// 结果入栈
// Obtain the data type of map1
// Can be used for different processing according to different types in the future
// Then push the result to the stack
typeOf $push $map1

// 输出类型名称
// Output the type name
pln 类型是： $pop

// 输出map1的内容
// Output the content of map1
plo $map1

// 获取map1中的键名为子女的项
// 结果放入变量list1中
// Obtain the key named child in map1
// Place the results in variable list1
getMapItem $list1 $map1 子女

// 获取list1中序号为0的项
// 结果放入变量map2中
// Get the item with sequence number 0 in list1
// Place the results in variable map2
getItem $map2 $list1 #i0

// 获取map2中键名为姓名的项
// 结果压栈
// Obtain items with key name in map2
// Then push the result to the stack
getMapItem $push $map2 姓名

// 输出弹栈值
// Output stack value
pln 姓名： $pop


```

运行后得到：

The running result:

```shell
类型是： map[string]interface {}
(map[string]interface {})map[姓名:张三 子女:[map[姓名:张胜利 年龄:5]] 年龄:39]
姓名： 张胜利
```

注意，typeOf指令可用于获取任意变量的数据类型名称，这在很多需要根据类型进行处理的场景下非常有用。typeOf获取到的值类型与宿主语言Go语言的一致，可参考Go语言的文档。

Note that the typeOf instruction can be used to obtain the data type name of any variable, which is very useful in many scenarios that require processing based on type. The value type obtained by typeOf is consistent with the host language Golang, and can refer to the Golang documentation.

&nbsp;

##### - **加载外部模块**（Loading external module）

&nbsp;

谢语言可以动态加载外部的代码文件并执行，这是一个很方便也很重要的功能。一般来说，我们可以把一些常用的、复用程度高的功能写成快速函数或一般函数放在单独的谢语言源代码文件中，然后在需要使用的代码中动态加载它们并使用其中的函数。可以构建自己的公共代码库，或者形成功能模块。

Xielang can dynamically load external code files and execute them, which is a very convenient and important feature. Generally speaking, we can write some commonly used and highly reusable functions as fast or general functions in a separate Xielang source code file, and then dynamically load them and use the functions in the code that needs to be used. You can build your own public code library or form functional modules.

下面的例子演示的是在一个代码文件中先后载入两个外部模块文件并调用其中的函数。

The following example demonstrates loading two external module files in a code file and calling the functions within them.

首先编写1个模块文件module1.xie，其中包含两个快速函数add1和sub1，功能很简单，就是两个数进行相加和相减。

Firstly, write a module file module1.xie, which contains two fast functions add1 and sub1. The function is very simple, which is to add and subtract two numbers.

*注意，由于快速函数与主函数共享全局变量空间，为避免冲突，建议变量名以大写的“L”结尾，以示只用于局部。另外还建议全局变量以大写的“G”结尾，一般的局部变量以大写的“T”结尾。这些不是强制要求，但也许能够起到一些避免混乱的效果。*

*Note that since the fast function shares the global variable space with the main function, to avoid conflicts, it is recommended to end the variable name with an uppercase "L" to indicate that it is only used locally. Additionally, it is recommended that global variables end with an uppercase 'G' and general local variables end with an uppercase 'T'. These are not mandatory requirements, but they may have some effect on avoiding confusion*

```go
:add1
    pop $v2L
    pop $v1L

    add $push $v1L $v2L

    fastRet

:sub1
    pop $v2L
    pop $v1L

    sub $push $v1L $v2L
    
    fastRet

```

然后再编写第二个模块文件module2.xie，其中包含一个普通函数mul1，作用是两个数相乘。

Then write the second module file module2.xie, which contains a regular function mul1 to multiply two numbers.

```go
:mul1
    var $v1
    var $v2
    var $outL

    getArrayItem $v1 $inputL 0
    getArrayItem $v2 $inputL 1

    mul $outL $v1 $v2

    ret


```

最后编写动态加载上面两个模块的例子代码（loadModule.xie）：

Finally, write example code for dynamically loading the above two modules (loadModule.xie):

```go
// 载入第1个代码文件module1.xie并压栈
// load code string from module file 1
// getCurDir $rs1
// joinPath $path1 $rs1 `scripts` `module1.xie`

loadText $code1 `scripts/module1.xie`

// 输出代码文件内容查看
// print the code string for reference
pln "code loaded: " "\n" $code1 "\n"

// 加载代码
// 并将结果值返回，成功将返回加载代码的第1行行号（注意是字符串类型）
// 失败将返回error对象表示的错误信息
// load the code to current VM
loadCode $rs $code1

// 检查是否出错，是则停止代码运行
// check if is error, terminate if true
checkErrX $rs

// 压栈两个整数
// push 2 values before fast-calling function
push #i11
push #i12

// 调用module1.xie文件中定义的快速函数add1
// fast-call the "add1" function defined in the file "module1.xie"
fastCall :add1

// 查看函数返回结果（不弹栈）
// print the result pushed into the stack from the function
// unlike pop, peek only "look" but not get the value out of the stack
plo $peek

// 再压入堆栈一个整数5
// push another value integer 5 into the stack
push #i5

// 调用module1.xie文件中定义的快速函数sub1
// fast-call the "sub1" function defined in the file "module1.xie"
fastCall :sub1

// 再次查看函数返回结果（不弹栈）
// print the result again
plo $peek

// 载入第2个代码文件module2.xie并置于变量code1中
// load text from another module file
loadText $code1 `scripts/module2.xie`

// 编译这个代码以节约一些后面载入的时间
// this time, compile it first(will save some time before running)
compile $compiledT $code1

checkErrX $compiledT

// 加载编译后的代码
// 由于不需要loadCode指令返回的结果，因此用$drop变量将其丢弃
// load the code and drop the result using the global variable $drop
loadCode $drop $compiledT

// there is a integer value 18 in the stack
// 此时栈中还有一个整数18

// 调用module2.xie文件中定义的一般函数mul1，并传入两个参数（整数99和弹栈值18）
// fast-call the "mul1" function defined in the file "module2.xie"
call $rs :mul1 #i99 $pop

// 查看函数返回结果
// print the result
plo $rs

// 退出程序执行
// 注意：如果不加exit指令，程序会继续向下执行module1.xie和module2.xie中的代码
// terminate the program
// if without the "exit" instruction here, the program will continue to run the code loaded by module1.xie and module2.xie
exit

```

代码中的重点是loadText指令和loadCode指令。loadText从指定路径读取纯文本格式的模块代码文件内容。loadCode文件则从字符串变量中读取代码并加载到当前代码的后面，如果成功，会返回这段代码的起始位置（注意是字符串格式），有些情况下会用到这个返回值。对于以函数为主的模块，在动态加载包含这些函数的文件后，就可以用call或fastCall指令来调用相应的函数了。

The focus in the code is on the loadText and loadCode instructions. The loadText instruction reads the content of the module code file in plain text format from the specified path. The loadCode instruction reads the code from a string variable and loads it after the current code. If successful, it returns the starting position of the code (note the string format), which may be used in some cases. For modules that primarily rely on functions, after dynamically loading the files containing these functions, the corresponding functions can be called using the call or fastCall instructions.

代码运行的结果是：

The running result is:

```shell
code loaded:  
 :add1
    pop $v2L
    pop $v1L

    add $push $v1L $v2L

    fastRet

:sub1
    pop $v2L
    pop $v1L

    sub $push $v1L $v2L
    
    fastRet 

(int)23
(int)18
(int)1782

```

&nbsp;

##### - **封装函数调用**（Sealed Function Call）

&nbsp;

封装函数与一般函数与快速函数的区别是：封装函数直接采用源代码形式调用，实际上会新启动一个谢语言虚拟机去执行函数代码，封闭性更好（相当于沙盒执行），也更灵活，参数和返回值通过堆栈传递；缺点是性能稍慢（因为要启动虚拟机并解析代码）。下面是封装函数调用的例子（sealCall.xie）：

The difference between sealed functions and general functions and fast functions is that sealed functions can be directly called in source code form, which actually launches a new Xielang virtual machine to execute the function code. It has better closure (equivalent to sandbox execution) and is more flexible, with parameters and return values passed through the stack; The disadvantage is that the performance is slightly slow (because the virtual machine needs to be started and the code needs to be compiled). The following is an example of sealed function calls (sealCall.xie):

```go
// 使用sealCall进行封装调用函数
// 与runCall不同，入参和出参是通过inputG和outG变量进行的
// 下面先准备两个用于调用函数的输入参数，其中第二个是采用堆栈存放
// Using sealCall to sealed function calls
// Unlike runCall, input and output parameters are carried out through inputG and outG variables
// Next, prepare two input parameters for calling the function, the second of which is stored in the stack
assign $a #f1.62
push #f2.8

// 封装函数调用会启动一个新虚拟机来运行代码、编译后对象或运行上下文
// 第一个参数是结果参数，不可省略，第二个参数为字符串形式的源代码或者编译后对象等
// 如果需要传入参数则从第三个参数开始，可以传入多个，因此inputG中存放的将是一个数组/列表
// Sealed function calls will start a new virtual machine to run code, compiled objects, or run context
// The first parameter is the result parameter and cannot be omitted. The second parameter is the source code or compiled object in string form, etc
// If parameters need to be passed in, starting from the third parameter, multiple can be passed in, so inputG will store an array/list
sealCall $rs `
    // inputG是一个数组/列表，其中包含所有输入参数
    // 使用getArrayItem指令从其中按索引获取所需的参数
    // inputG is an array/list contains all the input parameters
    // Use the getArrayItem instruction to retrieve the required parameters by index from it
    getArrayItem $num1 $inputG 0
    getArrayItem $num2 $inputG 1

    // 输出两个参数作为参考
    // output 2 values for reference
    pln num1= $num1
    pln num2= $num2

    // 将两个数相乘后将加过存入变量result
    // multiply 2 values and put the result to $result
    mul $result $num1 $num2

    // 输出结果变量参考
    // print the result value for reference
    pln $result

    // 封装函数将通过outG变量返回值
    // 如果要返回多个变量，可以使用数组/列表
    // return values in the global variable $outG
    // if more than one result, use array/list
    assign $outG $result
` $a $pop

// 输出函数返回值
// print the result from the function
pl "seal-function result: %v" $rs

```

代码中，封装函数直接用反引号扩起了多行的代码。sealCall指令需要一个参数指定结果变量，不可省略。第二个参数是字符串类型的源代码，也可以是编译后的对象，或者运行上下文，本例中传入了一个多行字符串，既是这个封装函数的代码。封装函数如果要返回值，需要使用全局变量outG返回，如果需要返回多个值，可以使用数组/列表，或者映射也可以。

In the code, the sealed function directly wraps multiple lines of code with back quotes. The sealCall instruction requires a parameter to specify the result variable and cannot be omitted. The second parameter is the source code of the string type, which can also be the compiled object or the runtime context. In this example, a multi line string is passed in, which is the code that encapsulates the function. If the sealed function needs to return a value, it needs to use the global variable outG to return it. If multiple values need to be returned, an array/list, or a mapping can also be used.

注意，封装函数可以以字符串形式的代码加载并执行的，这意味着封装函数也可以动态加载，例如从文件中读取代码后执行，这带来了很大的灵活性。另外，封装函数在单独的虚拟机中运行，和主函数的变量和堆栈空间都不冲突，因此可以编写更通用的函数。

Note that sealed functions can be loaded and executed as strings of code, which means that sealed functions can also be dynamically loaded, such as reading code from a file and executing it, which brings great flexibility. In addition, the sealed function runs in a separate virtual machine and does not conflict with the variables and stack space of the main function, so more general functions can be written.

上面代码的执行结果是：

The running result:

```shell
num1= 1.62
num2= 2.8
4.536
seal-function result: 4.536
```

sealCall指令如果有第三个以上的参数，将从第三个开始合并为数组传入新虚拟机中的inputG全局变量。

If the sealCall instruction has more than three parameters, it will be merged into an array starting from the third and passed into the inputG global variable in the new virtual machine.

&nbsp;

##### - **并发函数**（Concurrent function call）

&nbsp;

谢语言中的并发常用类似于封装函数的并发函数来实现，使用goCall/threadCall指令，这两个指令是等价的。下面是并发函数调用的例子（goCall.xie）：

Concurrency in Xielang is often implemented using concurrent functions similar to sealed functions, using the goCall/threadCall instruction, which is equivalent. The following is an example of concurrent function calls (goCall.xie):

```go
// 将变量a定义为一个指向任意类型数值的引用变量，并将其初始值赋为浮点数3.6
// 引用变量a指向的值将在线程中运行的并发函数内被修改
// Define variable a as a reference variable pointing to any type of numerical value, and assign its initial value to floating point 3.6
// The value pointed to by the reference variable a will be modified within the concurrent function running in the thread
var $a "*any" #f3.6

// 输出当前变量a本身及其指向的值作参考
// Output the current variable a itself and the value it points to as a reference
pl "a=%v, *a=%v" $a *$a

// 使用goCall指令调用并发函数
// 第一个参数是结果值，不可省略，并发函数的返回值仅表示函数是否启动成功，不代表其真正的返回值
// 第二个参数是字符串形式的并发函数代码，也可以是编译后对象或运行上下文
// 如果需要传递参数，从第三个参数开始可以传入多个参数，这些参数在函数体内可以通过inputG变量访问
// 由于可以有多个传入参数，inputG是一个数组/列表
// Calling concurrent functions using the goCall instruction
// The first parameter is the result value, which cannot be omitted. The return value of a concurrent function only indicates whether the function was started successfully and does not represent its true return value
// The second parameter is the concurrent function code in string form, which can also be a compiled object or runtime context
// If parameters need to be passed, multiple parameters can be passed in starting from the third parameter, which can be accessed within the function body through the inputG variable
// Due to the possibility of multiple input parameters, inputG is an array/list
goCall $rs `
    // 从inputG按索引顺序获取两个传入的参数
    // Obtain two incoming parameters in index order from inputG
    getArrayItem $arg1 $inputG 0
    getArrayItem $arg2 $inputG 1

    // 查看两个参数值
    // View two parameter values
    pln arg1= $arg1
    pln arg2= $arg2

    // 解引用第一个参数（即获取主函数中的引用变量a指向的值）
    // Dereference the first parameter (i.e. obtain the value pointed to by the reference variable a in the main function)
    unref $aNew $arg1

    // 输出变量a指向的值以供参考
    // Output the value pointed by variable a for reference
    pln "value in $a is:" $aNew

    // 无限循环演示不停输出时间
    // loop1是用于循环的标号
    // Infinite loop demonstration without stopping output time
    // Loop1 is a label used for the loop
    :loop1
        // 输出sub和变量arg2中的值
        // Output sub and the value in variable arg2
        pln sub $arg2

        // 获取当前时间并存入变量timeT
        // Obtain the current time and store it in the variable timeT
        now $timeT

        // 将timeT中的时间值赋给变量arg1指向的值
        // assignRef的第一个参数必须是一个引用变量
        // Assign the time value in timeT to the value pointed to by variable arg1
        // The first parameter of assignRef must be a reference variable
        assignRef $arg1 $timeT

        // 休眠2秒
        // Sleep for 2 seconds
        sleep #f2.0

        // 跳转到标号loop1（实现无限循环）
        // Jump to label loop1 (implementing infinite loop)
        goto :loop1
` $a "prefix"

// 主线程中输出变量a的值及其指向的值
// 变量名前加“*”表示取其指向的值，这时候一定是一个引用变量
// 此时刚开始启动并发函数，变量a中的值有可能还未改变
// The value of output variable a in the main thread and the value it points to
// Adding "*" before the variable name indicates taking the value it points to, which must be a reference variable
// At this point, the concurrent function has just started, and the value in variable a may not have changed yet
pln main $a *$a

// 注意，这里的标号loop1虽然与并发函数中的同名，但由于运行在不同的虚拟机中，因此不会冲突，可以看做是两个标号
// Note that although the label loop1 here has the same name as the concurrent function, it does not conflict as it runs on different virtual machines and can be considered as two labels
:loop1

    // 休眠1秒
    // Sleep for 1 second
    sleep #f1.0

    // 输出变量a中的值及其指向的值查看
    // 每隔一秒应该会变成新的时间
    // View the values in output variable a and the values it points to
    // Every second should become a new time
    pln a: $a *$a

    // 跳转到标号loop1（无限循环，可以用Ctrl+C键中止程序运行）
    // Jump to label loop1 (infinite loop, you can use Ctrl+C to abort program execution)
    goto :loop1
```

代码中有详细的注释，主线程中启动了一个子线程，也就是调用了并发函数，看看运行效果：

There are detailed comments in the code. A sub thread has been started in the main thread, which means calling a concurrent function. Let's see the running effect:

```shell
a=0xc000465250, *a=3.6
main 0xc000465250 3.6
arg1= 0xc000465250
arg2= prefix
value in $a is: 3.6
sub prefix
a: 0xc000465250 2023-04-17 10:14:21.46951 +0800 CST m=+0.025359501
sub prefix
a: 0xc000465250 2023-04-17 10:14:23.4820764 +0800 CST m=+2.037911601
a: 0xc000465250 2023-04-17 10:14:23.4820764 +0800 CST m=+2.037911601
sub prefix
a: 0xc000465250 2023-04-17 10:14:25.4853758 +0800 CST m=+4.041196801
a: 0xc000465250 2023-04-17 10:14:25.4853758 +0800 CST m=+4.041196801
sub prefix
a: 0xc000465250 2023-04-17 10:14:27.4892531 +0800 CST m=+6.045059801
^C

```

仔细观察程序的输出，可以看出并发函数中的输出每两秒1次，主线程中的输出每2秒一次，变量a中的值确实从最初的浮点数3.6到后来被并发函数变成了当前时间。

By carefully observing the output of the program, it can be seen that the output in the concurrent function is once every two seconds, and the output in the main thread is once every two seconds. The value in variable a has indeed changed from the initial floating point number of 3.6 to the current time by the concurrent function.

这个例子中也演示了对变量取引用与对引用解引用后取变量值的方法。

This example also demonstrates the methods of taking a reference to a variable and taking a variable value after removing the reference.

&nbsp;

##### - **用线程锁处理并发共享冲突**（Using Thread Locks to Handle Concurrent Sharing Conflicts）

&nbsp;

谢语言中，对于同时运行的几个线程间共享某个变量，对其进行读取和修改时可能产生的并发冲突问题，可以使用线程锁来控制解决。参看下面的例子（lock.xie）：

In Xielang, thread locks can be used to control and resolve concurrency conflicts that may arise when reading and modifying a variable shared between multiple threads running simultaneously. Please refer to the following example (lock.xie):

```go

// 给变量a赋值整数0
// 变量a将在线程中运行的并发函数中被修改
// Assign integer 0 to variable a
// Variable a will be modified in the concurrent function running in the thread
assign $a #i0

// 创建一个线程锁对象放入变量lock1中
// 指令new用于创建谢语言中一些基础数据类型或宿主语言支持的对象
// 除结果变量外第一个参数为字符串类型的对象名称
// lock对象是共享锁对象
// Create a thread lock object and place it in the variable lock1
// The instruction 'new' is used to create some basic data types or objects supported by the host language in Xielang
// The first parameter, except for the result variable, is an object type name of string type
// The lock object is a shared lock object
new $lock1 lock

// 定义一个并发函数体func1（用字符串形式定义）
// 并发函数如果使用goRunCall指令启动，会在新的运行上下文中运行
// 除了全局变量外，其中的变量名称不会与主线程冲突
// 因此可以传入变量引用或普通引用以便在并发函数中对其进行修改
// Define a concurrent function body func1 (defined in string form)
// If the concurrent function is started using the goRunCall instruction, it will run in a new running context
// Except for global variables, their names will not conflict with those in the main thread
// Therefore, variable contains a reference/pointer value or regular Xielang variable references can be passed in to modify them in concurrent functions
assign $func1 `
    // 获取两个传入的参数，参数是通过inputL变量传入的
    // 并发调用的函数体一般不需要返回参数
    // “[]”指令是getArrayItem指令的简写形式
    // Obtain two incoming parameters, which are passed in through the inputL variable
    // Function bodies that are called concurrently generally do not need to return parameters
    // The '[]' instruction is a shortened form of the getArrayItem instruction
    [] $arg1 $inputL 0
    [] $arg2 $inputL 1

    // 创建一个循环变量i并赋以初值0
    // Create a loop variable i and assign it an initial value of 0
    assign $i #i0

    // 无限循环演示不停将外部传入的变量a值加1
    // loop1是用于循环的标号
    // Infinite loop demonstration continuously increasing the value of variable a passed in externally by 1
    // Loop1 is a label used for loops
    :loop1
        // 调用传入的线程锁变量的加锁方法（lock）
        // 此处变量arg2即为外部传入的线程锁对象
        // 由于lock方法没有有意义的返回值，因此用内置变量drop将其丢弃
        // Call the lock method of the incoming thread lock variable
        // The variable arg2 here refers to the thread lock object passed in externally
        // Due to the lack of meaningful return values for the lock method, it is discarded using the built-in variable $drop
        method $drop $arg2 lock

        // 解引用变量a的引用，以便取得a中当前的值
        // Dereference the variable arg1 in order to obtain the current value in a
        unrefVar $aNew $arg1
    
        // 将其加1，结果放入变量result中
        // Add 1 to it and place the result in the variable result
        add $result $aNew #i1

        // 将变量arg1指向的变量（即a）中的值赋为result中的值
        // assignRefVar的第一个参数必须是一个引用
        // Assign the value in the variable (i.e. a) pointed to by variable arg1 to the value in result
        // The first parameter of assignRefVar must be a reference
        assignRefVar $arg1 $result

        // 调用线程锁的unlock方法将其解锁，以便其他线程可以访问
        // Call the unlock method of the thread lock to unlock it so that other threads can access it
        method $drop $arg2 unlock

        // 循环变量加1
        // Increase the loop variable by 1
        inc $i

        // 判断循环变量i的值是否大于或等于30
        // 即循环5000次
        // 判断结果值（布尔类型）放入变量r中
        // Determine whether the value of loop variable i is greater than or equal to 30
        // That is, 5000 cycles
        // Put the judgment result value (Boolean type) into the variable r
        >= $r $i #i30        


        // 如果r值为真（true），则转到标号beforeReturn处
        // If the value of r is true, go to the label beforeReturn
        if $r :beforeReturn

        // 休眠1秒钟
        // Sleep for 1 second
        sleep #f1.0

        // 跳转到标号loop1（实现无限循环）
        // Jump to label loop1 (implementing infinite loop)
        goto :loop1

    :beforeReturn
        // pass指令不进行任何操作，由于标号处必须至少有一条指令
        // 因此放置一条pass指令，实际上beforeReturn这里作用是结束线程的运行
        // 因为没有后续指令了
        // The pass instruction does not perform any operations as there must be at least one instruction at the label
        // Therefore, by placing a pass instruction, in fact, beforeReturn is used to end the thread's operation
        // Because there are no further instructions left
        pass
`

// 获取变量a的引用，存入变量p1中
// 将被传入并发函数中以修改a中的值
// Obtain a reference to variable a and store it in variable p1
// Will be passed into the concurrent function to modify the value in a
refVar $p1 $a

// 用goRunCall指令调用并发函数，结果参数意义不大，因此用$drop丢弃
// 第一个参数是字符串类型的函数体
// 后面跟随传入的两个参数
// 第一个传入参数是p1，即变量a的引用
// 第二个参数是线程锁对象，因为本身就是引用，因此可以直接传入
// Calling a concurrent function with the goRunCall instruction resulted in insignificant parameters, so $drop was used to discard them
// The first parameter is the function body of string type
// Following the two parameters passed in
// The first incoming parameter is p1, which is a reference to variable a
// The second parameter is the thread lock object, which is itself a reference and can be directly passed in
goRunCall $drop $func1 $p1 $lock1

// 再启动一个相同的线程
// Start another identical thread
goRunCall $drop $func1 $p1 $lock1

// 主线程中输出变量a的值
// 此时刚开始启动并发函数，变量a中的值有可能还未改变
// The value of output variable a in the main thread
// At this point, the concurrent function has just started, and the value in variable a may not have changed yet
pln main a= $a

// 注意，这里的标号loop1虽然与并发函数中的同名，但由于运行在不同的运行上下文中，因此不会冲突，可以看做是两个标号
// Note that although the label loop1 here has the same name as the concurrent function, it does not conflict as it runs in different running contexts and can be considered as two labels
:loop1

    // 休眠1秒
    // Sleep for 1 second
    sleep #f1.0

    // 输出变量a中的值查看
    // 由于同时启动了两个线程，并且都是每隔1秒将a中值加1
    // 因此每隔一秒输出的值会加2，最终达到60
    // 由于1秒时间点的细微差异，有时候也会是加3或加1
    // View the value in output variable a
    // Due to the simultaneous start of two threads, both of which increase the value of a by 1 every 1 second
    // Therefore, the output value will increase by 2 every second, ultimately reaching 60
    // Due to subtle differences in time points of 1 second, sometimes it can also be increased by 3 or 1
    pln main a= $a

    // 跳转到标号loop1（实现无限循环）
    // Jump to label loop1 (implementing infinite loop)
    goto :loop1

```

method指令用于调用对象的某个方法，这里是调用了线程锁的lock和unlock方法。method指令可以简写为mt。

The method instruction is used to call a method of an object, where the lock and unlock methods of the thread lock are called. The method instruction can be abbreviated as mt.

如果没有对线程锁对象加锁、解锁的操作（可以注释上其中method $drop $arg2 lock与unlock这两条语句尝试），程序运行的结果将是不确定的数字，每次都有可能结果不同，这是因为两个线程各自存取变量a中的值产生的冲突所致。例如，当第一个线程取到了a的值为10，在将其加1但还没有来得及把值（11）赋回给a的时候，第二个线程获取了当时的a值10，也将其加1后赋回给a，然后线程1再把11赋给a，这样虽然两个线程各执行了一个a=a+1的操作，但其实效果相当于只执行了1次。这样，我们如果再把线程中的休眠指令去掉，并增大结束条件到5000次，以便更好地显示出效果，最后程序结果应该是a的值小于理论值10000，还有一种可能是两个线程同时操作映射对象时使得程序崩溃。

If there is no lock or unlock operation on the thread lock object (you can comment on the two statements' method $drop $arg2 lock and unlock ' to try), the result of program execution will be an uncertain number, and the result may be different each time. This is due to conflicts between the values in variable a accessed by two threads. For example, when the first thread takes the value of a as 10 and adds it to 1 before it can assign the value (11) back to a, the second thread takes the value of a as 10 and adds it back to a, then thread 1 assigns 11 to a. Although each thread performs an operation with a=a+1, the effect is equivalent to only executing it once. In this way, if we remove the sleep instructions from the thread and increase the end condition to 5000 times to better display the effect, the final program result should be that the value of a is less than the theoretical value of 10000. Another possibility is that two threads operating on the mapped object simultaneously can cause the program to crash.

加上线程锁后，每次最终结果都将是准确的60（或5000次的话结果是10000），如下所示。

After adding thread locks, each final result will be an accurate 60 (or 10000 if 5000 times), as shown below.

```shell
main a= 0
main a= 2
main a= 4
main a= 6
main a= 8
main a= 10
main a= 12
main a= 14
main a= 16
main a= 18
main a= 20
main a= 22
main a= 24
main a= 26
main a= 28
main a= 30
main a= 32
main a= 34
main a= 36
main a= 40
main a= 40
main a= 42
main a= 44
main a= 47
main a= 48
main a= 50
main a= 52
main a= 54
main a= 56
main a= 58
main a= 60
main a= 60
main a= 60
main a= 60
main a= 60
main a= 60


```

&nbsp;

##### - **对象机制**(Object Model)

&nbsp;

谢语言提供一个通用的可扩展的对象机制，来提供集成宿主语言基本能力和库函数优势的方法，对象可以自行编写，可以使用宿主语言也可以使用谢语言本身编写（建设中），同时，谢语言也已经提供了一些内置的对象供直接使用。

Xielang provides a universal and extensible object mechanism to provide a method of integrating the basic capabilities of the host language and the advantages of library functions. Objects can be written on their own, either using the host language or using Xielanguage itself (under construction). At the same time, Xielang has also provided some built-in objects for direct use.

下面是使用内部对象string的一个例子(object.xie)，这个对象非常简单，仅仅封装了一个字符串，但提供了一些成员方法来对其进行操作，具体实现可以参考谢语言的源代码。

The following is an example of using an internal object string (object.xie). This object is very simple, only encapsulating a string, but providing some member methods to operate on it. For specific implementation, please refer to the source code of Xielang.

*注意，谢语言的对象一般包含本体值（例如string对象就是其包含的字符串）及可以调用的成员方法，还可能包含成员变量。*

*Note that Xielang objects generally contain inner values (such as string objects being the strings they contain), member methods that can be called, and may also contain member variables*

```go
// 新建一个string对象，赋以初值字符串“abc 123”，放入变量s中
// Create a new string object, assign the initial string 'abc 123', and place it in the variable 's'
newObj $s string `abc 123`

// 获取对象本体值，结果压栈
// Obtain the object's inner value and push the results into the stack
getObjValue $push $s

// 将弹栈值加上字符串“天气很好”，结果存入tmp
// Add the stack pop value with the string '天气很好' and store the result in $tmp
add $pop "天气很好"

// 输出tmp值供参考
// Output the value in $tmp for reference
pln $tmp

// 设置变量s中string对象的本体值为字符串“very”
// Set the inner value of the string object in variable s to the string 'very'
setObjValue $s "very"

// 输出对象值供参考
// Output the value in $tmp for reference
pln $s

// 调用该对象的add方法，并传入参数字符串“ nice”
// 该方法将把该string对象的本体值加上传入的字符串
// Call the add method of the object and pass in the parameter string ' nice'
// This method will add the inner value of the string object with the incoming string
callObj $s add " nice"

// 再次输出对象值供参考
// Output the value again
pln $s

// 调用该对象的trimSet方法，并传入参数字符串“ve”
// 该方法将把该string对象的本体值去掉头尾的字符v和e
// 直至头尾不是这两个字符任意之一
// Call the trimSet method of the object and pass in the parameter string 've'
// This method will remove the first and last characters v and e from the inner value of the string object
// Until the beginning and end are not either of these two characters
callObj $s trimSet "ve"

// 再次输出对象值供参考
// Output the value again
pln $s


```

代码运行的结果是：

The running result will be:

```shell
abc 123天气很好
very
very nice
ry nic
```

&nbsp;

##### - **快速/宿主对象机制**(Fast/Host Object Mechanism)

&nbsp;

谢语言也提供另一个new指令来实现快速的对象机制，也可以提供集成宿主语言基本能力和库函数优势的方法，对象使用上更简单。下面是一个例子（stringBuffer.xie），封装了一般语言中的可动态增长的字符串的功能。

Xielang also provides another new instruction to implement fast object mechanisms, as well as a method to integrate the basic capabilities of the host language and the advantages of library functions, making object usage simpler. Here is an example (stringBuffer.xie) that encapsulates the functionality of dynamically growing strings in other languages.

```go
// 本例演示快速/宿主对象机制
// 以及method/mt方法的使用、双引号与反引号是否转义等
// This example demonstrates the fast/host object mechanism
// And the use of method/mt methods, whether double and back quotes are escaped, etc

// strBuf即Go语言中的strings.Builder
// 是一个可以动态多次向其中添加字符串的缓冲区
// 最后可以一次性获取其中的所有内容为一个字符串
// StrBuf is the string. Builder in Go language
// It is a buffer that can dynamically add strings multiple times to it
// Finally, all the contents can be obtained as a string at once
new $bufT strBuf

// 调用bufT的append方法往其中写入字符串abc
// method（可以简写为mt）指令是调用对象的某个方法
// append/writeString/write方法实际上是一样的，都是向其中追加写入字符串
// 这里结果参数使用了$drop，因为一般用不到返回值
// Calling the append method of bufT to write the string abc to it
// The method (which can be abbreviated as mt) instruction is a method that calls an object's member function
// The append/writeString/write method is actually the same, appending and writing a string to the string buffer
// The result parameter here uses $drop, as the return value is generally not used
method $drop $bufT append abc


// 使用双引号括起的字符串中间的转义符会被转义
// The escape character in the middle of a string enclosed in double quotation marks will be escaped
method $drop $bufT writeString "\n"

mt $drop $bufT write 123

// 使用反引号括起的字符串中的转义符不会被转义
// Escape characters in strings enclosed in back quotes will not be escaped
mt $drop $bufT append `\n`

// 用两种方式输出bufT中的内容供参考
// Output the content in bufT in two ways for reference

// 调用bufT的str方法（也可以写作string、getStr等）获取其中的字符串
// Call the str method of bufT (which can also be written as string, getStr, etc.) to obtain the string in it
mt $rsT $bufT str

plo $rsT

// 直接用表达式来输出
// Directly using expressions to output
pln @`{mt $tmp $bufT str}`

```

运行输出：

The output is:

```shell
(string)"abc\n123\\n"
abc
123\n
```

&nbsp;

##### - **时间处理**(Time processing)

&nbsp;

谢语言中的时间处理的主要方式，直接参看下面的代码（time.xie）：

The main method of time processing in Xielang can be directly seen in the following code (time.xie):

```go
// 将变量t1赋值为当前时间
// #t后带空字符串或now都表示当前时间值
// Assign variable t1 to the current time
// Either an empty string or 'now' after # t indicates the current time value
assign $t1 #t

// 输出t1中的值查看
// View the value in output t1
plo $t1

// 用字符串表示时间
// “=”是指令assign的简写写法
// Using a string to represent time
// '=' is the abbreviation for the instruction 'assign'
= $t2 #t`2022-08-06 11:22:00`

pln t2= $t2

// 简化的字符串表示形式
// Simplified string representation
= $t3 #t`20220807112200`

pl t3=%v $t3

// 带毫秒时间的表示方法
// Representation method with millisecond time
= $t4 #t`2022-08-06 11:22:00.019`

pl t4=%v $t4

// 时间的加减操作
// 与时间的计算，如果有数字参与运算（除了除法之外），一般都是以毫秒为单位
// Addition and subtraction of time
// The calculation of time, if there are numbers involved in the operation (except for division), is generally in milliseconds
pl t2-3000毫秒=%v @`$t2 - 3000`

pl t2+50000毫秒=%v @`$t2 + 50000`

pl 当前时间+50000毫秒=%v @`{now} + 50000`

pl t3-t2=%v(毫秒) @`$t3 - $t2`

// 注意，如果不用括号，表达式计算将严格从左到右，没有运算符的优先级
// Note that if parentheses are not used, the expression will be evaluated strictly from left to right, without operator priority
pl t3-t2=%v(小时) @`($t3 - $t2) / #i1000 / #i60 / #i60`

// 时间的比较
// Comparison of time
pl `t2 < t3 ? %v` @`$t2 < $t3`

pl `t2 >= t3 ? %v` @`$t2 >= $t3`

pl `t4 == t3 ? %v` @`$t4 == $t3`

pl `t1 != t3 ? %v` @`$t1 != $t3`

// 用convert指令转换时间
// Convert time using the convert instruction
convert $tr `2021-08-06 11:22:00` time

pln tr= $tr

// 用convert指令将时间转换为普通字符串
// Convert time to a regular string using the convert instruction
convert $s1 $tr str

pln s1= $s1

// 用convert指令将时间转换为特定格式的时间字符串
// Convert time to a specific format time string using the convert instruction
convert $s2 $tr timeStr `2006/01/02_15.04.05`

pln s2= $s2

// 用convert指令将时间转换为UNIX时间戳格式
// Convert time to UNIX timestamp format using the convert instruction
convert $s3 $tr tick

pln s3= $s3

// 用convert指令将UNIX格式时间戳转换为时间
// Convert UNIX format timestamp to time using the convert instruction
convert $t5 `1628220120000` time

pln t5= $t5

// UTC相关
// 用convert指令转换时间为UTC时区
// UTC related
// Convert time to UTC time zone using the convert instruction
convert $trUTC `2021-08-06 11:22:00` time -global

pln trUTC= $trUTC

nowUTC $t6

pln t6= $t6

timeToLocal $t7 $t6

pln t7= $t7

timeToGlobal $t8 $t7

pln t8= $t8

// 用var指令也可以定义一个时间类型变量
// 默认值是当前时间
// Using the var instruction can also define a time type variable
// The default value is the current time
var $t9 time

// 调用时间类型变量的addDate方法将其加上1个月
// 三个参数分别表示要加的年、月、日，可以是负数
// 结果还放回t9
// Call the addDate method of a time type variable to add 1 month to it
// The three parameters represent the year, month, and day to be added, which can be negative numbers
// Return the result to t9
mt $t9 $t9 addDate 0 1 0

// 调用时间类型变量的format函数将其格式化为字符串
// 格式参数参考[这里](https://pkg.go.dev/time#pkg-constants)
// Call the format function of a time type variable to format it as a string
// Format parameter reference [here]（ https://pkg.go.dev/time#pkg -constants)
mt $result $t9 format "20060102"

// 应输出 t9: 20220825
// Will output t9: 20220825
pl "t9: %v" $result

```

运行后输出为：

Output is like:

```shell
(time.Time)2022-07-25 09:22:03.918723645 +0800 CST m=+0.014649799
t2= 2022-08-06 11:22:00 +0800 CST
t3=2022-08-07 11:22:00 +0800 CST
t4=2022-08-06 11:22:00.019 +0800 CST
t2-3000毫秒=2022-08-06 11:21:57 +0800 CST
t2+50000毫秒=2022-08-06 11:22:50 +0800 CST
当前时间+50000毫秒=2022-07-25 09:22:53.919167824 +0800 CST m=+50.015093974
t3-t2=86400000(毫秒)
t3-t2=24(小时)
t2 < t3 ? true
t2 >= t3 ? false
t4 == t3 ? false
t1 != t3 ? true
tr= 2021-08-06 11:22:00 +0800 CST
s1= 2021-08-06 11:22:00 +0800 CST
s2= 2021/08/06_11.22.00
s3= 1628220120000
t5= 2021-08-06 11:22:00 +0800 CST
trUTC= 2021-08-06 11:22:00 +0000 UTC
t6= 2022-07-25 01:22:03.919574126 +0000 UTC
t7= 2022-07-25 09:22:03.919574126 +0800 CST
t8= 2022-07-25 01:22:03.919574126 +0000 UTC
t9: 20220825
```

&nbsp;

##### - **错误处理**(Error Handling)

&nbsp;

谢语言一般使用一个简化的错误处理模式，参看下面的代码（onError.xie）：

Xielang generally uses a simplified error handling mode, as shown in the following code (onError.xie):

```go

// 设置错误处理代码块为标号handler1处开始的代码块
// onError指令后如果不带参数，表示清空错误处理代码块
// set error handler to the code block at label :handler1
// onError instruction with no parameters will clear the defined error handlers
onError :handler1

// 故意获取一个超出数组长度索引的结果，将产生运行时异常
// trigger an error on purpose
var $array1 array
getArrayItem $item $array1 1

// 此处代码正常情况应执行不到
// 但错误处理代码块将演示如何返回此处继续执行
// the code below will not run normally
// but the error handler will redirect to this label
:next1

// 输出一个提示信息
// output a piece of message for reference
pln "calculation completed(and the error handler)"

// 退出程序
// terminate the program
exit

// 错误处理代码块
// error handler
:handler1
    // 发生异常时，谢语言将会将出错时出错代码的行号、错误提示信息和详细代码运行栈信息分别存入全局变量$lastLineG, $errorMessageG, $errorDetailG
    // 错误处理代码块可以根据这几个值进行相应显示或其他处理
    // the error info will in 3 global variables: $lastLineG, $errorMessageG, $errorDetailG
    // error handler can handle them

   // 输出错误信息
    // output the message
    pl "error occurred while running to line %v: %v, detail: %v" $lastLineG $errorMessageG $errorDetailG

    // 跳转到指定代码位置继续执行
    // jump to the specific position(label)
    goto :next1

```

关键点是使用onError指令，它带有一个参数，一般是如果代码运行发生异常时将要跳转到的错误处理代码块的标号。onError指令既是指定代码运行错误时，用于处理错误的代码块。这样，如果代码运行发生任何运行时错误，谢语言将会将出错时出错代码的行号、错误提示信息和详细代码运行栈信息分别存入全局变量$lastLineG, $errorMessageG, $errorDetailG，然后从该标号处开始执行。错误处理代码块一般需要根据几个出错信息进行提示或其他相应处理，最后可以选择跳转到指定位置执行，或者终止程序运行等操作。还有一种常用的处理方式是跳转到出错行号的下一个行号处继续执行。


The key point is to use the onError instruction, which takes a parameter and is usually the label of the error handling code block to jump to if an exception occurs during code execution. The onError instruction is both a code block used to handle errors when specifying code execution errors. In this way, if any runtime errors occur during code execution, Xielang will store the line number, error prompt information, and detailed code runtime stack information of the error code into the global variables \$lastLineG, \$errorMessageG, and \$errorDetailG, respectively, and then execute from that label. Error handling code blocks generally require prompts or other corresponding processing based on several error messages. Finally, you can choose to jump to the specified location for execution, or terminate program execution and other operations. Another commonly used processing method is to jump to the next line number where the error occurred and continue execution.

本段代码的运行结果是：

The running result will be:

```shell
error occurred while running to line 10: runtime error, detail: (Line 10: getArrayItem $item $array1 1) index out of range: 1/0
calculation completed(and the error handler)
```

谢语言中的错误处理也可以使用类似Go语言异常处理的defer指令来处理，主要用于确保打开的文件、网络连接、数据库连接等最终会被关闭，具体参见下面一节。

Error handling in Xielang can also be handled using the defer instruction similar to Go language exception handling, mainly used to ensure that open files, network connections, database connections, etc. will eventually be closed. Please refer to the following section for details.

&nbsp;

##### - **延迟执行指令 defer**

&nbsp;

与Go语言类似，谢语言也支持用defer指令延迟执行另一条指令，即如果在一条指令前加上defer指令，则该条指令不会立即执行，执行的时机是主程序退出前或函数执行退出前，参看下面的代码（defer.xie）：

Similar to Golang, Xielang also supports using the defer instruction to delay the execution of another instruction. If a defer instruction is added before an instruction, the instruction will not execute immediately, and the execution time is before the main program exits or the function exits. Please refer to the following code (defer.xie):

```go
// 延迟执行指令1
// defer指令可以后接一个字符串表示的指令
// defer instruction 1
// The defer instruction can be followed by an instruction represented by a string
defer `pl "main defer: %v" test1`

// 延迟执行指令2
// defer指令遵循“后进先出”的规则，即后指定的defer指令将先被执行
// defer指令后也可以直接跟一条指令而不是字符串
// defer instruction 2
// the deferred instructions will be running by order(first in last out) when the function returns or the program exits, or error occurrs
// The defer instruction can also be directly followed by an instruction instead of a string
defer pl "main defer: %v" test2

// deferStack $pln
// exit

// defer指令也可以后接一个编译好的代码
// The defer instruction can also be followed by a compiled code
compile $code1 `
pln main defer 3 in compiled code piece
pln "..."
`

defer $code1

// 百分号引导的参数表示将代码编译
// The parameter guided by the percentage sign means to compile the code
defer %`
pln main defer 4 in compiled code piece
pln "___"
`

pln 1

// 函数中的延迟执行
// call a function to test defer instruction in functions
call $rs :func1

pln func1 returns $rs

exit

:func1
    defer `pl "sub defer: %v" test1`

    pln sub1

    // 故意做一个会出现错误的指令，这里是除零操作
    // trigger an error on purpose
    eval $r1 `#i10 / #i0`

    // 检查出错则中断程序，此时应执行本函数内的defer和主函数内的defer
    // check if error occurred, and since it is, the deferred instructions defined in this function and up to root function of the VM will be run
    checkErrX $r1

    // 下面的代码不会被执行到
    // code below will never be reached
    pln "10/0=" $r1

    ret $r1

```

可以看出，defer指令也可以被用在异常/错误处理的场景下。

It can be seen that the defer instruction can also be used in exception/error handling scenarios.

&nbsp;

##### - **关系数据库访问**（Relational Database Access）

&nbsp;

谢语言主程序支持常见的关系型数据库的访问与操作，直接参看下面访问SQLite3数据库的代码例子（sqlite.xie）：

The main program of Xielang supports the access and operation of common relational database. Please refer to the following code example for accessing SQLite3 database (sqlite.xie):

```go
// 本例演示谢语言对SQLite 3库的创建与增删改查
//This example demonstrates the creation, addition, deletion, and modification of SQLite 3 library by Xielang

// 判断是否已存在该库（SQLite库是放在单一的文件中的）
// 注意请确保c:\tmp文件夹已存在
// 结果放入变量b中
// Determine if the library already exists (SQLite library is placed in a single file)
// Please ensure that the c:\tmp folder already exists
// Place the results in variable b
fileExists $b `c:\tmpx\test.db`

// 如果否则跳到下一步继续执行
// 如果存在则删除该文件
// removeFile指令的运行结果将被丢弃（因为使用了内置全局变量drop）
// If not, skip to the next step to continue execution
// If present, delete the file
// The result of the removeFile instruction will be discarded (due to the use of the built-in global variable drop)
ifNot $b :next
	removeFile $drop `c:\tmpx\test.db`

:next1
// 创建新库
// dbConnect用于连接数据库
// 除结果参数外第一个参数是数据库驱动名称，支持sqlite3、mysql、godror（即Oracle）、mssql（即MS SQLServer）等
// 第二个参数是连接字符串，类似 server=129.0.3.99;port=1433;portNumber=1433;user id=sa;password=pass123;database=hr 或 user/pass@129.0.9.11:1521/testdb 等
// SQLite3的驱动将基于文件创建或连接数据库
// 所以第二个参数直接给出数据库文件路径即可
// Create a new library
// dbConnect is used to connect to a database
// The first parameter besides the result parameter is the database driver name, which supports sqlite3, mysql, godror (i.e. Oracle), msql (i.e. MS SQLServer), etc
// The second parameter is the connection string, similar to server=129.0.3.99; port=1433; portNumber=1433; user id=sa; password=pass123; Database=hr or user/ pass@129.0.9.11 : 1521/testdb et al
// The driver of SQLite3 will create or connect databases based on files
// So the second parameter directly provides the database file path
dbConnect $db "sqlite3" `c:\tmpx\test.db`

// 判断创建（或连接）数据库是否失败
// rs中是布尔类型表示变量db是否是错误对象
// 如果是错误对象，errMsg中将是错误原因描述字符串
// Determine if the creation (or connection) of the database has failed
// Is the Boolean type in rs indicating whether the variable db is the wrong object
// If it is an error object, errMsg will contain the error reason description string
isErr $rs $db $errMsg

// 如果为否则继续执行，否则输出错误信息并退出
// If not, continue executing, otherwise output an error message and exit
ifNot $rs :next2
	pl "创建数据库文件时发生错误：%v" $errMsg
	exit

:next2

// 将变量sqlStmt中放入要执行的建表SQL语句
// Place the variable sqlStmt into the table building SQL statement to be executed
assign $sqlStmt = `create table TEST (ID integer not null primary key, CODE text);`

// 执行SQL语句，dbExec用于执行insert、delete、update等SQL语句
// Execute SQL statements, dbExec is used to execute SQL statements such as insert, delete, update, etc
dbExec $rs $db $sqlStmt

// 判断是否SQL执行出错，方式与前面连接数据库时类似
// Determine if there was an SQL execution error, similar to when connecting to the database earlier
isErr $errStatus $rs $errMsg

ifNot $errStatus :next3
	pl "执行SQL语句建表时发生错误：%v" $errMsg

	// 出现错误时，因为数据库连接已打开，因此需要关闭
	// When an error occurs, it needs to be closed because the database connection is already open
	dbClose $drop $db

	exit

:next3

// 进行循环，在库中插入5条记录
// i是循环变量
// Loop and insert 5 records into the library
// I is a cyclic variable
assign $i #i0

:loop1
assign $sql `insert into TEST(ID, CODE) values(?, ?)`

// genRandomStr指令用于产生随机字符串
// The genRandomStr instruction is used to generate random strings
genRandomStr $str1

dbExec $rs $db $sql $i $str1

isErr $errStatus $rs $errMsg

ifNot $errStatus :next4
	pl "执行SQL语句新增记录时发生错误：%v" $errMsg
	dbClose $drop $db

	exit

:next4
inc $i
< $i #i5
if $tmp :loop1

// 进行数据库查询，验证查看刚刚新增的记录
// Perform database queries, verify and view the newly added records
assign $sql `select ID, CODE from TEST`

// dbQuery指令用于执行一条查询（select）语句
// 结果将是一个数组，数组中每一项代表查询结果集中的一条记录
// 每条记录是一个映射，键名对应于数据库中的字段名，键值是相应的字段值，但均转换成字符串类型
// The dbQuery instruction is used to execute a select statement
// The result will be an array, where each item represents a record in the query result set
// Each record is a mapping, and the key name corresponds to the field name in the database. The key value is the corresponding field value, but it is converted to a string type
dbQuery $rs $db $sql

// dbClose指令用于关闭数据库连接
// The dbClose directive is used to close a database connection
dbClose $drop $db

pln $rs

// 用toJson指令将结果集转换为JSON格式以便输出查看
// Convert the result set to JSON format using the toJson directive for output viewing
toJson $jsonStr $rs -indent -sort

pln $jsonStr


```

执行结果是（确保c:\tmpx目录已经存在）：

The running result is (ensuring that the c:\tmpx directory already exists):

```shell
[map[CODE:YRKOEt ID:0] map[CODE:moODkc ID:1] map[CODE:we7Ey9 ID:2] map[CODE:fF7dRd ID:3] map[CODE:9X6KAu ID:4]]
[
  {
    "CODE": "YRKOEt",
    "ID": "0"
  },
  {
    "CODE": "moODkc",
    "ID": "1"
  },
  {
    "CODE": "we7Ey9",
    "ID": "2"
  },
  {
    "CODE": "fF7dRd",
    "ID": "3"
  },
  {
    "CODE": "9X6KAu",
    "ID": "4"
  }
]
```

可以看到，数据库中被新增了5条记录，并查询成功。

It can be seen that 5 new records have been added to the database and successfully queried.

常见类型的数据库连接字符串组合如下：

The common types of database connection string combinations are as follows:

- SQLite： 
  ``` dbConnect $db "sqlite3" `c:\tmpx\test.db` ```

- MySQL： 
  ``` dbConnect $db "mysql" `user:pass@tcp(192.168.1.27:3306)/dbname` ```

- MSSQL/SQL Server： 
  ``` dbConnect $db "sqlserver" "server=192.168.100.10;port=1433;portNumber=1433;user id=userName;password=userPass;database=DB1;encrypt=disable" ```

- Oracle： 
  ``` dbConnect $db "godror", "user/pass@db.somedomain.com:1521/ORCL" ```

- Oracle（另一种驱动 another driver）： 
  ``` dbConnect $db "oracle", "oracle://user:pass@db.somedomain.com:1521/ORCL" ```

&nbsp;

##### - **微服务/应用服务器**（Microservices/Application Server）

&nbsp;

谢语言主程序自带一个服务器模式，支持一个轻量级的WEB/应用/API三合一服务器。可以用下面的命令行启动：

Xielang main program comes with a server mode that supports a lightweight WEB/Application/API 3 in 1 server. You can start it using the following command line:

```shell
D:\tmp>xie -server -dir=scripts
[2023/05/24 08:40:35] Xie micro-service framework V1.2.2 -port=:80 -sslPort=:443 -dir=scripts -webDir=scripts -certDir=.
[2023/05/24 08:40:35] starting https service on port :443...
starting http service on port :80...
[2023/05/24 08:40:35] failed to start https service: open server.crt: The system cannot find the file specified.
```

可以看到，谢语言的服务器模式可以用-server参数启动，并可以用-port参数指定HTTP服务端口（注意加冒号），用-sslPort指定SSL端口，-certDir用于指定SSL服务的证书文件目录（应为server.crt和server.key两个文件），用-dir指定服务的根目录，-webDir用于指定静态页面和资源的WEB服务。这些参数均有默认值，不输入任何参数可以看到。

As can be seen, the server mode of Xielang can be started with the '-server' parameter, and the '-port' parameter can be used to specify the HTTP service port (please add a colon), '-sslPort' can be used to specify the SSL port, '-certDir' can be used to specify the certificate file directory of the SSL service (which should be two files: server.crt and server.key), '-dir' can be used to specify the root directory of the service, and '-webDir' can be used to specify the web service for static pages and resources. These parameters have default values and can be seen without entering any parameters.

输出信息中的错误是因为没有提供SSL证书，SSL服务将启动不了，加上证书就可以了。

The error in the output information is because the SSL certificate was not provided, and the SSL service will not be able to start. Adding the certificate files will suffice.

此时，用浏览器访问本机的 http://127.0.0.1:80 就可以访问一个谢语言编写的网页服务了。

At this point, access the local http://127.0.0.1:80 You can access a web service written in Xielang.

假设在指定的目录下包含 xmsIndex.xie、xmsTmpl.html、xmsApi.xie三个文件，可以展示出谢语言建立的应用服务器支持的各种模式。

Assuming that the specified directory contains three files: xmsIndex.xie, xmsTmpl.html, and xmsApi.xie, various modes supported by the application server established by Xielang can be displayed.

首先用浏览器访问 http://127.0.0.1/xmsTmpl.html ，这将是访问一般的WEB服务，因为WEB目录默认与服务器根目录相同，所以将展示根目录下的xmsTmpl.html这个静态文件，也就是一个例子网页。

First, access it with a browser http://127.0.0.1/xmsTmpl.html This will be accessing general web services, as the web directory defaults to the same as the server root directory. Therefore, the static file xmsTmpl.html under the root directory will be displayed, which is an example web page.

![截图](http://xie.topget.org/example/xie/snap/snap1.jpg)

可以看到，该网页文件中文字“请按按钮”后的“{{text1}}”标记，这是我们后面展示动态网页功能时所需要替换的标记。xmsTmpl.html文件的内容如下：

You can see that the "{{text1}}" tag after the text "请按按钮（Please press the button）" in the webpage file is the tag that we need to replace when displaying the dynamic webpage function later. The content of the xmsTmpl.html file is as follows:

```html
<html>
<body>
    <script>
        function test() {
            let xhr = new XMLHttpRequest();

            xhr.open('POST', 'http://127.0.0.1:80/xms/xmsApi', true);

            xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded")

            xhr.onload = function(){
                alert(xhr.responseText);
            }

            xhr.send("param1=abc&param2=123");
        }
    </script>

    <div>
        <span>请按按钮{{text1}}：</span><button onclick="javascript:test();">按钮1</button>
    </div>

</body>
</html>
```

然后我们尝试进行动态网页输出，也就是类似PHP、ASP或其他类似的架构支持的后台动态渲染网页的方式。访问 http://127.0.0.1/xms/xmsIndex ，URL中加上xms路径，这是一个虚拟路径，表示服务器将去根目录下寻找xmsIndex.xie文件来执行，该代码将输出网页内容。我们来看下xmsIndex.xie文件的内部。

Then we try to perform dynamic web page output, which is a way of dynamically rendering web pages in the background supported by PHP, ASP, or other similar architectures. visit http://127.0.0.1/xms/xmsIndex Add the xms path to the URL, which is a virtual path indicating that the server will search for the xmsIndex.xie file in the root directory to execute. This code will output the webpage content. Let's take a look inside the xmsIndex.xie file.

```go
// 设定默认的全局返回值变量outG为字符串TX_END_RESPONSE_XT
// 默认谢语言服务器如果收到处理请求的函数返回结果是TX_END_RESPONSE_XT
// 将会终止处理，否则将把返回值作为字符串输出到网页上
// Set the default global return value variable outG to the string TX_END_RESPONSE_XT
// If the default Xielang server receives a function to process a request, the return result is TX_END_RESPONSE_XT
// Processing will be terminated, otherwise the return value will be output as a string to the webpage
assign $outG "TX_END_RESPONSE_XT"

// 获得相应的网页模板
// joinPath指令将把多个文件路径合并成一个完整的文件路径
// 第一个参数表示结果将要放入的变量，这里的$push表示压栈
// basePathG是内置全局变量，表示服务的根目录
// Obtain the corresponding webpage template
// The joinPath directive will merge multiple file paths into one complete file path
// The first parameter represents the variable to be placed in the result, where $push represents stack pushing
// basePathG is a built-in global variable that represents the root directory of the service
joinPath $push $basePathG `xmsTmpl.html`

pln $basePathG
pln $peek

// 将该文件作为文本字符串载入，结果压栈
// Load the file as a text string and stack the results
loadText $push $pop

// 替换其中的{{text1}}标记为字母A
// Replace the {{text1}} marked with the letter A
strReplace $push $pop "{{text1}}" "A"

// 将弹栈值写网页输出
// responseG也是内置的全局变量，表示要写入的网页输出对象
// Write the stack value to the webpage output
// ResponseG is also a built-in global variable that represents the webpage output object to be written to
writeResp $responseG $pop

// 终止请求响应处理微服务
// Terminate Request Response Processing Microservice
exit

```

谢语言服务器模式中，每一个http请求都将单开一个虚拟机进行处理，可以看做一个微服务的概念。例子中的微服务仅仅是将载入的网页模板中的指定标记替换掉然后输出到网页，虽然简单，但已经展现出了动态网页的基本原理，即能够在输出网页前进行必要的、可控的渲染。

In Xielang server model, each HTTP request will be processed by a separate virtual machine, which can be seen as a microservice concept. The microservice in the example only replaces the specified tags in the loaded webpage template and outputs them to the webpage. Although it is simple, it has already demonstrated the basic principle of dynamic webpage, that is, it can perform necessary and controllable rendering before outputting the webpage.

我们访问 http://127.0.0.1/xms/xmsIndex 这个网址（或者叫URL路径），将会看到如下结果：

We visited http://127.0.0.1/xms/xmsIndex This website address (or URL path) will result in the following results:

![截图](http://xie.topget.org/example/xie/snap/snap2.jpg)

可以发现原来的标记确实被替换成了大写的字母A，验证了动态网页的效果。

It can be found that the original tag has indeed been replaced with the uppercase letter A, verifying the effect of dynamic web pages.

再看上面的网页模板文件xmsTmpl.html，其中的按钮点击后将执行JavaScript函数test，其中进行了一个AJAX请求，然后将请求的结果用alert函数输出出来。这是一个典型的客户端访问后台API服务的例子，我们来看看如何实现这个后台API服务。下面是也在服务器根目录下的xmsApi.xie文件中的内容：

Looking at the webpage template file xmsTmpl. html above, once the button is clicked, the JavaScript function test will be executed, where an AJAX request is made and the result of the request will be output using the alert function. This is a typical example of a client accessing a backend API service. Let's take a look at how to implement this backend API service. The following is the content of the xmsApi.xie file also located in the server root directory:

```go
// 获取当前时间放入变量t
// Get the current time and put it into variable t
nowStr $t

// 输出参考信息
// 其中reqNameG是内置全局变量，表示服务名，也就是访问URL中最后的部分
// paraMapG也是全局变量，表示HTTP请求包含的URL参数或Form参数（可以是GET请求或POST请求中的）
// Output reference information
// Where reqNameG is a built-in global variable that represents the service name, which is the last part of the access URL
// paraMapG is also a global variable that represents the URL or Form parameters contained in HTTP requests (which can be in GET or POST requests)
pl `[%v] %v args: %v` $t $reqNameG $paraMapG

// 设置输出响应头信息（JSON格式）
// Set output response header information (JSON format)
setRespHeader $responseG "Content-Type" "text/json; charset=utf-8"

// 写响应状态为整数200（HTTP_OK），表示是成功的请求响应
// The write response status is an integer of 200 (HTTP_oK), indicating a successful request response
writeRespHeader $responseG #i200

// 用spr指令拼装响应字符串
// Assembling response strings using spr instructions
spr $str1 "请求是：%v，参数是：%v" $reqNameG $paraMapG

// 用genJsonResp生成封装的JSON响应，也可以自行输出其他格式的字符串
// Generate encapsulated JSON responses using genJsonResp, or output strings in other formats on your own
genJsonResp $respStr $requestG "success" $str1

// 将响应字符串写输出（到网页）
// Write and output the response string (to a webpage)
writeResp $responseG $respStr

// 结束处理函数，并返回TX_END_RESPONSE_XT以终止响应流的继续输出
// End processing function and return TX_END_RESPONSE_XT to terminate the continued output of the response stream
exit TX_END_RESPONSE_XT

```

这样，我们如果点击网页中的按钮1，会得到如下的alert弹框：

In this way, if we click button 1 on the webpage, we will get the following alert pop-up:

![截图](http://xie.topget.org/example/xie/snap/snap4.jpg)

这是因为网页xmsTmpl.html中，通过AJAX访问了 http://127.0.0.1:80/xms/xmsApi 这个服务，而我们的谢语言服务器会寻找到xmsApi.xie（自动加上了.xie文件名后缀）并执行，因此会输出我们希望的内容。

This is because the webpage xmsTmpl. html is accessed through AJAX http://127.0.0.1:80/xms/xmsApi This service, and our Xielang server will find xmsApi.xie (automatically added with the .xie file name suffix) and execute it, so it will output the content we want.

至此，一个麻雀虽小五脏俱全的WEB/应用/API多合一服务器的例子就完整展现出来了，已经足够一般小型的应用服务，并且基本无外部依赖，部署也很方便，只需一个主程序以及拷贝相应目录即可。

At this point, an example of a sparrow's small and versatile WEB/application/API multi in one server has been fully demonstrated. It is already sufficient for a general and small application service, and has almost no external dependencies. Deployment is also very convenient, only requiring a main program and copying the corresponding directory.

&nbsp;

##### - **网络（HTTP）客户端**（Network(HTTP) Client）

&nbsp;

用谢语言实现一个网络客户端也非常容易，以上面的网络服务端为例，访问这些服务的客户端代码（httpClient.xie）如下：

It is also very easy to implement a network client using Xielang. Taking the network server as an example, the client code (httpClient.xie) for accessing these services is as follows:

```go
// getWeb指令可以用于各种基于HTTP的网络请求，
// 此处是获取某URL处的网页内容
// 第一个参数pageT用于存放访问的结果内容
// -timeout参数用于指定超时时间，单位是秒
// The getWeb directive can be used for various HTTP based network requests,
// This is to obtain the webpage content at a certain URL
// The first parameter pageT is used to store the accessed result content
// The - timeout parameter is used to specify the timeout time, in seconds
getWeb $pageT "http://127.0.0.1/xms/xmsIndex" -timeout=15

// 输出获取到的内容参考
// Output obtained content for reference
pln $pageT

// 定义一个映射类型的变量mapT
// 用于存放准备POST的参数
// Define a variable of mapping type mapT
// Used to store parameters for preparing for POST
var $mapT map

// 设置示例的POST参数
// Set POST parameters for the example
setMapItem $mapT param1 value1
setMapItem $mapT param2 value2

// 输出映射内容参考
// Output the map object for reference
pln $mapT

// 以POST的方式来访问WEB API
// getWeb指令除了第一个参数必须是返回结果的变量，
// 第二个参数是访问的URL，其他所有参数都是可选的
// method还可以是GET等
// encoding用于指定返回信息的编码形式，例如GB2312、GBK、UTF-8等
// headers是一个JSON格式的字符串，表示需要加上的自定义的请求头内容键值对
// 参数中可以有一个映射类型的变量或值，表示需要POST到服务器的参数
// Accessing WEB APIs through POST
// The getWeb instruction must be a variable that returns the result, except for the first parameter,
// The second parameter is the URL to access, all other parameters are optional
// The method can also be GET, etc
// Encoding is used to specify the encoding format of the returned information, such as GB2312, GBK, UTF-8, etc
// Headers is a JSON formatted string that represents the custom request header content key value pairs that need to be added
// There can be a mapping type variable or value in the parameter that represents the parameter that needs to be POST to the server
getWeb $resultT "http://127.0.0.1:80/xms/xmsApi" -method=POST -encoding=UTF-8 -timeout=15 -headers=`{"Content-Type": "application/json"}` $mapT

// 查看结果
// View the result
pln $resultT

```

示例中演示了直接获取网页和用POST形式访问API服务的方法，运行效果如下：

The example demonstrates the method of directly obtaining web pages and accessing API services through POST, and the running effect is as follows:

```shell
<html>
<body>
    <script>
        function test() {
            let xhr = new XMLHttpRequest();

            xhr.open('POST', 'http://127.0.0.1:80/xms/xmsApi', true);

            xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded")

            xhr.onload = function(){
                alert(xhr.responseText);
            }

            xhr.send("param1=abc&param2=123");
        }
    </script>

    <div>
        <span>请按按钮A：</span><button onclick="javascript:test();">按钮1</button>
    </div>

</body>
</html>
map[param1:value1 param2:value2]
{"Status":"success","Value":"请求是：xmsApi，参数是：map[param1:value1 param2:value2]"}
```

可以看到，程序顺利获得到了所需的服务器响应。

You can see that the program successfully obtained the required server response.

&nbsp;

##### - **手动编写Api服务器**（Manually writing Api servers）

&nbsp;

谢语言也支持自己手动编写各种基于HTTP的服务器，下面是一个API服务器的例子（apiServer.xie）：

Xielang also supports manually writing various HTTP based servers. Below is an example of an API server (apiServer.xie):

```go
// 新建一个路由处理器
// Create a new routing processor
newMux $muxT

// 设置处理路由“/test”的处理函数
// 第4个参数是字符串类型的处理函数代码
// 将以新的虚拟机运行
// 虚拟机内将默认有4个全局变量：
// requestG 表示http请求对象
// responseG 表示http响应对象
// paraMapG 表示http请求传入的query参数或post参数
// inputG 是调用setMuxHandler指令传入的第3个参数的值
// Set the processing function for processing route '/test'
// The 4th parameter is the code for the string type processing function
// Will run as a new virtual machine
// There will be 4 global variables by default within the virtual machine:
// requestG represents the HTTP request object
// responseG represents the HTTP response object
// paraMapG represents the query or post parameters passed in by HTTP requests
// inputG is the value of the third parameter passed in by calling the setMuxHandler instruction
setMuxHandler $muxT "/test" #i123 `

// 输出参考信息
// Output reference information
pln "/test" $paraMapG

// 拼装输出的信息字符串
// spr类似于其他语言中的sprintf函数
// Assembly output information string
// Spr is similar to the sprintf function in other languages
spr $strT "[%v] 请求名: test，请求参数： %v，inputG：%v" @'{nowStr}' $paraMapG $inputG

// 设置输出的http响应头中的键值对
// Set the key value pairs in the output HTTP response header
setRespHeader $responseG "Content-Type" "text/json; charset=utf-8"

// 设置输出http响应的状态值为200（表示成功，即HTTP_OK）
// Set the status value of the output HTTP response to 200 (indicating success, i.e. HTTP_oK)
writeRespHeader $responseG 200

// 准备一个映射对象用于拼装返回结果的JSON字符串
// Prepare a mapping object for assembling JSON strings that return results
var $resMapT map

setMapItem $resMapT "Status" "success"
setMapItem $resMapT "Value" $strT

// 转换为JSON
// Convert to JSON
toJson $jsonStrT $resMapT

// 写http响应内容，即前面拼装并转换的变量jsonStrT中的JSON字符串
// Write the HTTP response content, which is the JSON string in the variable jsonStrT that was previously assembled and converted
writeResp $responseG $jsonStrT

// 设置函数返回值为TX_END_RESPONSE_XT
// 此时响应将中止输出，否则将会把该返回值输出到响应中
// Set the function return value to TX_ END_ RESPONSE_ XT
// At this point, the response will stop any output, otherwise the return value will be output to the response
assign $outG  "TX_END_RESPONSE_XT"

`

pl "启动服务器……(请用浏览器访问 http://127.0.0.1:8080/test 查看运行效果)"
// pl "Start the server... (Please use a browser to access http://127.0.0.1:8080/test to view the running effect)"

// 在端口8080上启动http服务器
// 指定路由处理器为muxT
// 结果放入变量resultT中
// 由于startHttpServer如果执行成功是阻塞的
// 因此resultT只有失败或被Ctrl-C中断时才会有值
// Start the HTTP server on port 8080
// Specify the routing processor as muxT
// Place the results in the variable resultT
// Due to the fact that startHttpServer is blocked if executed successfully
// Therefore, resultT only has a value when it fails or is interrupted by Ctrl-C
startHttpServer $resultT ":8080" $muxT

```

运行后，用浏览器访问下面的网址进行测试：

After running, use a browser to access the following website for testing:

```
http://127.0.0.1:8080/test?param1=abc&param2=123
```

可以看到网页中会显示类似下面的JSON格式的输出：

You can see that the webpage will display output in JSON format similar to the following:

```
{
  "Status": "success",
  "Value": "[2022-05-17 15:11:57] 请求名: test，请求参数： map[param1:abc param2:123]，inputG：123"
}
```

当然，一般API服务都是用编程的形式而非浏览器访问，用浏览器比较适合做简单的测试。

Of course, most API services are accessed through programming rather than a browser, browsers are more suitable for simple testing.

&nbsp;

##### - **静态WEB服务器**（Static WEB server）

&nbsp;

谢语言实现静态WEB服务器则更为简单，见下例（webServer.xie）：

Implementing a static WEB server using Xielang is simpler, as shown in the following example (webServer.xie):

```go
// 新建一个路由处理器
// Create a new routing processor
newMux $muxT

// 设置处理路由“/static/”后的URL为静态资源服务
// 第3个参数是对应的本地文件路径
// 例如：访问 http://127.0.0.1:8080/static/basic.xie
// 而当前目录是c:\tmp，那么实际上将获得c:\tmp\scripts\basic.xie
// Set the URL after processing route '/static/' as a static resource service
// The third parameter is the corresponding local file path
// For example: accessing http://127.0.0.1:8080/static/basic.xie
// And the current directory is c:\tmp, so in reality, you will get c:\tmp scripts basic.xie
setMuxStaticDir $muxT "/static/" "./scripts" 

setMuxStaticDir $muxT "/" "." 

pln 启动服务器……
// pln "Starting the server..."


// 在端口8080上启动http服务器
// 指定路由处理器为muxT
// 结果放入变量resultT中
// 由于startHttpServer如果执行成功是阻塞的
// 因此resultT只有失败或被Ctrl-C中断时才会有值
// Start the HTTP server on port 8080
// Specify the routing processor as muxT
// Place the results in the variable resultT
// Due to the fact that startHttpServer is blocked if executed successfully
// Therefore, resultT only has a value when it fails or is interrupted by Ctrl-C
startHttpServer $resultT ":8080" $muxT

```

运行后，访问http://127.0.0.1:8080/static/basic.xie，将获得类似下面的结果：

After running, access http://127.0.0.1:8080/static/basic.xie , results similar to the following will be obtained:

```
// 本例演示做简单的加法操作
// This example demonstrates performing a simple addition operation

// 将变量x赋值为浮点数1.8
// Assign variable x to floating point 1.8
assign $x #f1.8

// 将变量x中的值加上浮点数2
// 结果压入堆栈
// Add the value in variable x to floating point number 2
// Result pushed onto the stack
add $push $x #f2

// 将堆栈顶部的值弹出到变量y
// Pop the value at the top of the stack onto the variable y
pop $y

// 将变量x与变量y中的值相加，结果压栈
// Add the values of variable x and variable y, and the result is stacked
add $push $x $y

// 弹出栈顶值并将其输出查看
// pln指令相当于其他语言中的println函数
// Pop up the top value of the stack and view its output
// The pln instruction is equivalent to the println function in other languages
pln $pop

// 脚本返回一个字符串“10”
// 如果有全局变量$outG声明过，则将其作为脚本返回值返回，谢语言主程序会将其输出
// The script returns a string of '10'
// If a global variable $outG has been declared, it will be returned as a script return value, and the Xielang main program will output it
= $outG 10

```

实际上读取了当前目录的scripts子目录下的basic.xie文件展示。

Actually, the server read the basic.xie file in the scripts subdirectory of the current directory and show it as the browsing result.

&nbsp;

##### - **动态网页服务器**（Dynamic Web Server）

&nbsp;

如果想要实现动态网页服务器，类似PHP、JSP、ASP等，可以参考之前的微服务/应用服务器和手动编写API服务器等例子，很容易实现。

If you want to implement a dynamic web server, such as PHP, JSP, ASP, etc., you can refer to previous examples of microservices/application servers and manually writing API servers, which are easy to implement.

&nbsp;

##### - **博客系统**（Implementing a tiny blog system）

&nbsp;

谢语言内置已经具备能力实现一个简单的博客系统。博客系统对比一般的网站服务器，主要需要增加下面几个功能：

Xielang already has the ability to implement a simple blog system built-in. Compared to general website servers, the blog system mainly needs to add the following functions:

- 支持注册、登录与鉴权
- 支持编辑文章
- 支持将特定格式的文章渲染成网页以便展示

- Support registration, login, and authentication
- Support for editing articles
- Support for rendering articles in specific formats into web pages for display

下面我们就举例说明谢语言实现一个最简单博客系统的方法。

Below, we will provide an example to illustrate the method of implementing the simplest blog system using Xielang.

首先，我们先建立一个登录服务（登录页面此处略去，有了登录服务接口，登录网页很容易就可以实现）。登录服务的目的是在用户成功登录以后获取一个令牌（token），此后在需要令牌鉴权的时候（例如编辑文章）会需要将该令牌传入。例子（blog/xms/xlogin.xie）如下：

Firstly, we will establish a login service (omitted here on the login page, with the login service interface, it is easy to implement logging into the webpage). The purpose of the login service is to obtain a token after the user successfully logs in, and then when token authentication is required (such as editing an article), the token will need to be passed in. An example is as follows:

```go
// 新建一个字符串缓冲区用于调试输出
// 实际系统中可取消此功能
// Create a new string buffer for debugging output
// This function can be cancelled in the actual system
new $debufG strBuf

// 跳转到主函数处执行
// Jump to the main function for execution
goto :main

// 几个出错时的处理分支
// 每个分支一般对应一种错误类型
// Several processing branches in case of errors
// Each branch generally corresponds to one error type
:fail1
    // 拼装错误信息字符串
    // Assembly error message string
    spr $tmps "empty %v" $fail1Reason

    // 生成JSON格式的错误对象
    // Generate JSON formatted error object
    genResp $result $requestG "fail" $tmps

    // 将其写入到HTTP请求响应中
    // Write it into the HTTP request response
    writeResp $responseG $result

    // 退出整个HTTP请求响应脚本
    // Exit the entire HTTP request response script
    exit

:fail2
    spr $tmps "require SSL"

    genResp $result $requestG "fail" $tmps

    writeResp $responseG $result

    exit

:fail3
    spr $tmps "%v" $fail1Reason

    genResp $result $requestG "fail" $tmps

    writeResp $responseG $result

    exit

// 通用错误处理函数
// General error handling function
:handler1
    spr $failMsg "internal error(line %v): %v（%v）" $lastLineG $errorMessageG $errorDetailG

    genResp $result $requestG "fail" $failMsg debuf $debufG

    writeResp $responseG $result

    exit

// 主函数开始
// Start of main function
:main

// 设定错误处理函数
// 所有未经处理的错误都将转入该函数进行处理
// Set error handling function
// All unhandled errors will be transferred to this function for processing
onError :handler1

// 默认返回TX_END_RESPONSE_XT，表示HTTP响应将不再输出（除了本脚本中已经输出的之外）
// Default return TX_END_RESPONSE_XT, indicating that the HTTP response will no longer be output (except for the content already output in this script)
= $outG "TX_END_RESPONSE_XT"

// 设置响应头中的字段
// Set the fields in the response header
setRespHeader $responseG "Content-Type" "text/json; charset=utf-8"

// 设置响应头状态为200，表示HTTP响应成功
// Set the response header status to 200, indicating successful HTTP response
writeRespHeader $responseG #i200

pl "[%v] %v params: %v" @'{nowStr}' $reqNameG $paraMapG

// 获取HTTP请求中的URL、协议等信息，可用于强制SSL判断等场景
// Obtain URL, protocol, and other information from HTTP requests, which can be used to enforce SSL judgment and other scenarios
mb $urlT $requestG URL

# pl urlT:%#v $urlT

mb $schemeT $requestG Scheme

# pln schemeT $schemeT

mb $protoT $requestG Proto

# pln protoT $protoT

mb $tlsT $requestG TLS

isNil $tlsT

// 打开下面的注释将强制要求此请求通过https访问
// if $tmp :fail2

# plv $requestG

// 获取请求参数中的appCode，u，p和secret，分别表示应用代码、用户名、密码和密钥
// 可以通过这些参数来进行鉴权，决定是否要授予令牌
// 本例中只通过判断密码来进行令牌授权
// 密钥是用于加密令牌的，可以为空值，将使用默认密钥
// Obtain the appCode, u, p, and secret in the request parameters, representing the application code, username, password, and key, respectively
// These parameters can be used for authentication to determine whether to grant a token
// In this example, token authorization is only performed by determining the password
// The key is used for encrypting tokens and can be empty, and then the default key will be used
getMapItem $appCode $paraMapG app

writeStr $drop $debufG @'"appCode: " + $appCode'

= $fail1Reason appCode

if @`$appCode == $undefinedG` :fail1

getMapItem $user $paraMapG u

= $fail1Reason user

if @`{isUndef $tmp $user}` :fail1

getMapItem $password $paraMapG p ""

= $fail1Reason password

if @`($password == "")` :fail1

= $fail1Reason "password not match"

if @`($password != "123456")` :fail3

getMapItem $secret $paraMapG secret ""

writeStr $_ $debufG @`(" -secret=" + $secret)`

// 生成令牌
// ifThenElse指令相当于JavaScript语言中的三元操作符（类似 a?true:false)
// Generate Token
// The ifThenElse instruction is equivalent to a ternary operator in JavaScript language (similar to a? True: false)
genToken $result $appCode $user admin @'{ifThenElse ($secret == "") "" ("-secret=" + $secret)}'

// 生成HTTP请求的响应字符串，JSON格式
// Generate a response string for HTTP requests in JSON format
genResp $result $requestG "success" $result debuf $debufG

// 写入HTTP响应
// Write HTTP response
writeResp $responseG $result

exit

```

然后，我们启动谢语言服务器，假设所需文件都在/goprjs/src/github.com/topxeq/xie/cmd/scripts/blog目录下：

Then, we start the Xielang server, assuming that the required files are all in the/goprjs/src/github.com/topxeq/xie/cmd/scripts/blog directory:

```shell
xie -server -port=:80 -sslPort=:443 -dir=/goprjs/src/github.com/topxeq/xie/cmd/scripts/blog/xms -webDir=/goprjs/src/github.com/topxeq/xie/cmd/scripts/blog/web -certDir=/goprjs/src/github.com/topxeq/xie/cmd/scripts/blog/cert -verbose
```

获取令牌的方法如下（为了安全起见，代码可以限制必须用https访问，另外参数最好使用POST方式传递，这里为了演示方便，采用了GET方式）：

The method to obtain a token is as follows (for security reasons, the code can restrict access to HTTPS, and it is best to use POST to pass parameters. Here, for demonstration purposes, GET method is used):

```
http://127.0.0.1/xms/xlogin?app=app1&u=userName&p=123456&secret=sdf789
```

其中，app是应用名称，可以自己设定，u是用户名，p是密码，secret是令牌加密秘钥（可以省略）。返回信息类似下面：

Among them, app is the application name, which can be set by oneself. u is the username, p is the password, and secret is the token encryption key (which can be omitted). The return information is similar to the following:

```
{
  "Status": "success",
  "Value": "9DCA7F736D56758385877E8A6E628D92727F848B7D81534E4B554F614943595E56635867",
  "debug": ""
}
```

Value字段中是后面可用的令牌。

The Value field contains the tokens available later.

然后，我们来架设博客服务。以Linux服务器为例，假定我们在/mnt/xms实现我们的博客服务，我们以下面的命令启动谢语言服务器：

Then, let's set up a blog service. Taking a Linux server as an example, assuming we implement our blog service in/mnt/xms, we start the Xielang server with the following command:

```shell
xie -server -port=:80 -sslPort=:443 -dir=/mnt/xms -webDir=/mnt/web -certDir=/mnt/cert -verbose
```

此时/mnt/web下为我们的静态网页文件，/mnt/xms下为我们的动态网页文件（前面的xlogin.xie也应该放在这个目录下），SSL证书放在/mnt/cert（因为server.crt和server.key两个文件）。一个特殊的约定是，/mnt/xms目录下的doxc.xie文件默认为博客处理的代码文件，访问 http://blog.example.com/xc/test 这样的请求（路径xc是预设的虚拟路径）时，将被交给doxc.xie来处理。因此我们根据自己需要修改该文件即可，一个典型例子如下：

At this point,/mnt/web is our static webpage file, and/mnt/xms is our dynamic webpage file(The previous xlogin.xie should also be placed in this directory). The SSL certificate is placed in/mnt/cert (because server.crt and server.key are two files). A special convention is that the doxc.xie file in the/mnt/xms directory defaults to the code file processed by the blog http://blog.example.com/xc/test When such a request is made, it will be handed over to doxc.xie for processing. Therefore, we can modify the file according to our own needs, a typical example is as follows:

```go

// 设置默认返回值为TX_END_RESPONSE_XT以避免多余的网页输出
// Set the default return value to TX_END_RESPONSE_XT to avoid unnecessary web page output
= $outG "TX_END_RESPONSE_XT"

pl "[%v] %v params: %v" @'{nowStr}' $reqNameG $paraMapG

// 设定错误和提示页面的HTML，其中的TX_main_XT等标记将被替换为有意义的内容
// Set HTML for error and prompt pages, where TX_ main_ Marks such as XT will be replaced with meaningful content
= $infoTmpl `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="content-type" content="text/html; charset=UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title></title>
</head>
<body>
    <div style="text-align: center;">
        <div id="main" style="witdth: 60%; margin-top: 3.0em; font-weight: bold; font-size: 2.0em; color: TX_mainColor_XT;">
            TX_main_XT
        </div>
        <div id="info" style="witdth: 90%; margin-top: 3.0em; font-size: 1.5em;">
            TX_info_XT
        </div>
    </div>
</body>
</html>
`

// 下面将放置一些快速调用的函数，因此这里直接跳转到main标号执行主程序代码
// Below will be some quick calling functions, so here we will directly jump to the main label to execute the main program code
goto :main

// 用于输出错误提示页面的函数
// Function for outputting error prompt pages
:fatalReturn
    getErrStrX $errStrL $pop
    strReplace $result $infoTmpl TX_info_XT $errStrL

    strReplace $result $result TX_main_XT $pop

    strReplace $result $result TX_mainColor_XT "#FF1111"

    writeResp $responseG $result

    exit

// 用于输出信息提示页面的函数
// Functions for outputting information prompt pages
:infoReturn

    strReplace $result $infoTmpl TX_info_XT $pop

    strReplace $result $result TX_main_XT $pop

    strReplace $result $result TX_mainColor_XT "#32CD32"

    writeResp $responseG $result

    exit

// 主函数代码入口
// Main function code entry
:main

// 新建一个字符串缓冲区（即可变长字符串）用于输出调试信息
// Create a new string buffer (i.e. a variable length string) for outputting debugging information
new $debufG strBuf

// reqNameG预设全局变量中存放的是请求路由
// 例如，访问http://example.com/xc/h/a1
// 则reqNameG为h/a1
// 将其分割为h和a1两段
// The reqNameG preset global variable stores the request routing
// For example, accessing http://example.com/xc/h/a1
// Then reqNameG is h/a1
// Divide it into two segments: h and a1
strSplit $listT $reqNameG "/" 2

// 加入调试信息
// Add debugging information
mt $drop $debufG append $listT

// 获取子请求的第一部分，即子请求名称（本例中为h）
// Obtain the first part of the sub request, which is the sub request name (in this case, h)
getItem $subReqT $listT 0

// 获取子请求的第二部分，即子请求参数，常用于表示对应资源的路径（本例中为a1）
// Obtain the second part of the sub request, which is the sub request parameter, commonly used to represent the path of the corresponding resource (in this example, a1)
getItem $subReqArgsT $listT 1

pln subReqT: $subReqT

// 如果子请求名称为h，则表示以网页形式输出页面（路径由子请求参数指定）
// If the sub request name is h, it means outputting the page as a web page (the path is specified by the sub request parameter)
ifEval `$subReqT == "h"` :+1 :next1

    setRespHeader $responseG "Content-Type" "text/html; charset=utf-8"

    writeRespHeader $responseG #i200

    strTrim $relDirT $subReqArgsT

    // 获取文件绝对路径，至于变量absPathT中
    // basePathG是启动谢语言服务器时指定的根目录
    // 例如，如果启动谢语言服务器时指定的根路径是 /mnt/xms，请求是：http://example.com/xc/h/a1 
    // 则实际输出的文件是 /mnt/xms/pages/a1.html
    // Obtain the absolute path of the file, ant put it in the variable absPathT
    // basePathG is the root directory specified when starting the Xielang server
    // For example, if the root path specified when starting the Xielang server is /mnt/xms, the request is: http://example.com/xc/h/a1 
    // The actual output file is/mnt/xms/pages/a1.html
    joinPath $absPathT $basePathG "pages" $relDirT

    pln absPathT: $absPathT

    // 如果子请求参数后缀不是“.html”或“.htm”，则加上后缀“.html”
    // If the subrequest parameter suffix is not ".html" or ".htm", add the suffix ".html"
    strEndsWith $b2T $absPathT ".html" ".htm"

    if $b2T :inext5
        + $absPathT $absPathT ".html"

    :inext5
    // 读取文件内容
    // Read File Content
    loadText $fcT $absPathT

    ifErrX $fcT :+1 :+2
        // fastCall后跳转标号后的参数将被依次压栈
        // The parameters after the jump label after fastCall will be sequentially pushed onto the stack
        fastCall :fatalReturn "action failed" $fcT

    // 将文件内容写入到HTTP响应
    // Write the file content to the HTTP response
    writeResp $responseG $fcT

    // 结束HTTP请求响应
    // End the HTTP request response
    exit

:next1

// 如果子请求名称为edit则表示编辑该页面
// 由于编辑操作一般需要权限验证，因此需要URL参数中传递通过xlogin接口获取的token
// 例如，需要这样访问： http://example.com/xc/edit/a1.html?token=96A4617B681F8668667971817C57767C73828C4D38304D47474E5153493958544F
// 注意，这里的文件名后缀不可省略
// If the sub request name is edit, it means editing the page
// Due to the fact that editing operations typically require permission verification, tokens obtained through the xlogin interface need to be passed in the URL parameters
// For example, it is necessary to access: http://example.com/xc/edit/a1.html?token=96A4617B681F8668667971817C57767C73828C4D38304D47474E5153493958544F
// Note that the file name suffix cannot be omitted here
ifEval `$subReqT == "edit"` :+1 :next2
    setRespHeader $responseG "Content-Type" "text/html; charset=utf-8"

    writeRespHeader $responseG #i200

    // 先检查token
    // secret是获取token时约定的密钥
    // Check the token first
    // Secret is the key agreed upon when obtaining the token
    getMapItem $tokenT $paraMapG "token"

    checkToken $r0 $tokenT -sercret=sdf789

    isErrX $r1 $r0 $msgT

    if $r1 :+1 :+2
        fastCall :fatalReturn "auth failed" $msgT
    
    pln token: $r0

    strSplit $list1T $r0 "|"

    getItem $userNameT $list1T 1

    // 只允许用户名为admin的用户操作
    // Only users with username admin are allowed to operate
    == $userNameT "admin"

    if $tmp :inext2
        fastCall :fatalReturn "auth failed" "user not exists"

    // 获取文件绝对路径
    // Obtain the absolute path of the file
    :inext2
    strTrim $relDirT $subReqArgsT

    joinPath $absPathT $basePathG "pages" $relDirT

    pln absPathT: $absPathT

    // 获取post参数ta1，如果存在则表示是保存操作
    // Obtain the post parameter ta1. If it exists, it means it is the save action

    getMapItem $ta1T $paraMapG ta1

    isUndef $b1T $ta1T

    if $b1T :inext4 
        // 保存文件

        extractFileDir $fileDirT $absPathT

        ensureMakeDirs $rs1T $fileDirT

        ifErrX $rs1T :+1 :+2
            fastCall :fatalReturn "failed to create path" $rs1T

        saveText $rs2T $ta1T $absPathT

        ifErrX $rs2T :+1 :+2
            fastCall :fatalReturn "failed to save file" $rs2T

    // 读取原有文件并展示
    // Read the original file and display it
    :inext4
    ifFileExists $b1 $absPathT

    = $fcT ""

    ifNot $b1 :+2
        loadText $fcT $absPathT

    // 编辑页面的HTML模板
    // HTML template for the edit page
    = $editTmplT `
    <!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="content-type" content="text/html; charset=UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title></title>
<script>
    window.onload = function() {
        // 设置文本输入框中TAB键的处理
        // Set the TAB key handler in the text input box
        document.getElementById('ta1').onkeydown = function(e) {
            if (e.keyCode == 9) {
                e.preventDefault();
                var indent = "\t";
                var start = this.selectionStart;
                var end = this.selectionEnd;
                var selected = window.getSelection().toString();
                selected = indent + selected.replace(/\n/g, '\n' + indent);
                this.value = this.value.substring(0, start) + selected
                        + this.value.substring(end);
                this.setSelectionRange(start + indent.length, start
                        + selected.length);
            }
        };
    }
</script>
</head>
<body>
<div id="div1" style="text-align: center; width: 100%; height: 100%;">
    <div style="width: 60%; margin: 0 auto; font-weight: bold; font-size: 2.0em;">
        <p>TX_filePath_XT</p>
    </div>
    <form method="POST">
        <div id="main" style="width: 80%; margin: 0 auto; height: 100%;">
            <textarea id="ta1" name="ta1" style="width: 100%; height: 30em; font-size: 1.5em;">TX_textAreaValue_XT</textarea>
        </div>
        <div style="width: 60%; margin: 0 auto; font-weight: bold; font-size: 2.0em;">
            <button type="submit">Save</button>
        </div>
    </form>
</div>
</body>
</html>
    `

    // 直接作为HTML代码插入页面，则需要进行HTML编码
    // Directly inserting the page as HTML code requires HTML encoding
    htmlEncode $rs1 $fcT

    strReplace $rs2 $editTmplT TX_textAreaValue_XT $rs1
    strReplace $rs2 $rs2 TX_filePath_XT $relDirT

    writeResp $responseG $rs2

    exit

:next2

// 如果子请求名称为t，则表示以纯文本形式输出页面
// If the sub request name is t, it means that the page is output in plain text form
ifEval `$subReqT == "t"` :+1 :next3
    strTrim $relDirT $subReqArgsT

    joinPath $absPathT $basePathG "pages" $relDirT

    // 如果路径不以“.txt”结尾，自动加上后缀“.txt”
    // If the path does not end with ". txt", automatically add the suffix ".txt"
    strEndsWith $b2T $absPathT ".txt"

    if $b2T :+2
        + $absPathT $absPathT ".txt"

    loadText $fcT $absPathT

    ifErrX $fcT :+1 :+2
        fastCall :fatalReturn "action failed" $fcT

    setRespHeader $responseG "Content-Type" "text/plain; charset=utf-8"

    writeRespHeader $responseG #i200

    writeResp $responseG $fcT

    exit

:next3

// 如果子请求名称为md，则表示以markdown形式渲染后输出页面
// If the sub request name is md, it means that the page is rendered as a markup and output
ifEval `$subReqT == "md"` :+1 :next4
    strTrim $relDirT $subReqArgsT

    joinPath $absPathT $basePathG "pages" $relDirT

    pln absPathT: $absPathT

    // 如果路径不以“.md”结尾，自动加上后缀“.md”
    // If the path does not end with ". md", automatically add the suffix ". md"
    strEndsWith $b2T $absPathT ".md"

    if $b2T :inext3
        + $absPathT $absPathT ".md"

    :inext3
    loadText $fcT $absPathT

    isErrX $errT $fcT $msgT

    if $errT :+1 :+2
        fastCall :fatalReturn 操作失败 $msgT

    renderMarkdown $fcT $fcT

    setRespHeader $responseG "Content-Type" "text/html; charset=utf-8"

    writeRespHeader $responseG #i200

    writeResp $responseG $fcT
    exit

:next4

// 如果子请求名称为editxms，则表示编辑谢语言代码
ifEval `$subReqT == "editxms"` :+1 :next5

    setRespHeader $responseG "Content-Type" "text/html; charset=utf-8"

    writeRespHeader $responseG #i200

    getMapItem $tokenT $paraMapG "token"

    checkToken $r0 $tokenT

    isErrX $r1 $r0 $msgT

    if $r1 :+1 :+2
        fastCall :fatalReturn "auth failed" $msgT
    
    pln token: $r0

    strSplit $list1T $r0 "|"

    getItem $userNameT $list1T 1

    == $userNameT "admin"

    if $tmp :inext6
        fastCall :fatalReturn "auth failed" "user not exists"

    // 获取文件绝对路径
    :inext6
    strTrim $relDirT $subReqArgsT

    joinPath $absPathT $basePathG x $relDirT

    pln absPathT: $absPathT

    // 获取post参数ta1，如果存在则表示是保存

    getMapItem $ta1T $paraMapG ta1

    isUndef $push $ta1T

    if $pop :inext7 
        // 保存文件

        extractFileDir $push $absPathT

        ensureMakeDirs $push $pop

        isErrX $errT $pop $msgT

        if $errT :+1 :+2
            fastCall :fatalReturn "failed to create path" $msgT

        saveText $push $ta1T  $absPathT

        isErrX $errT $pop $msgT

        if $errT :+1 :+2
            fastCall :fatalReturn "failed to save file" $msgT

    // 读取原有文件并展示
    :inext7
    ifFileExists $b1 $absPathT

    = $fcT ""

    ifNot $b1 :+2
        loadText $fcT $absPathT

    = $editTmplT `
    <!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="content-type" content="text/html; charset=UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title></title>
<script>
window.onload = function() {
    // 设置文本输入框中TAB键的处理
    // Set the TAB key handler in the text input box
    document.getElementById('ta1').onkeydown = function(e) {
        if (e.keyCode == 9) {
            e.preventDefault();
            var indent = "\t";
            var start = this.selectionStart;
            var end = this.selectionEnd;
            var selected = window.getSelection().toString();
            selected = indent + selected.replace(/\n/g, '\n' + indent);
            this.value = this.value.substring(0, start) + selected
                    + this.value.substring(end);
            this.setSelectionRange(start + indent.length, start
                    + selected.length);
        }
    };
}
</script>
</head>
<body>
<div id="div1" style="text-align: center; width: 100%; height: 100%;">
    <div style="width: 60%; margin: 0 auto; font-weight: bold; font-size: 2.0em;">
        <p>TX_filePath_XT</p>
    </div>
    <form method="POST">
        <div id="main" style="width: 80%; margin: 0 auto; height: 100%;">
            <textarea id="ta1" name="ta1" style="width: 100%; height: 30em; font-size: 1.5em;">TX_textAreaValue_XT</textarea>
        </div>
        <div style="width: 60%; margin: 0 auto; font-weight: bold; font-size: 2.0em;">
            <button type="submit">Save</button>
        </div>
    </form>
</div>
</body>
</html>
    `

    htmlEncode $rs1 $fcT

    strReplace $rs2 $editTmplT TX_textAreaValue_XT $rs1
    strReplace $rs2 $rs2 TX_filePath_XT $relDirT

    writeResp $responseG $rs2

    exit
 
:next5

// 如果不是任何已知的子请求名称，转到提示页面
// If it is not any known sub request name, go to the prompt page
fastCall :infoReturn "unknown request" $subReqT

exit


```

运行后，先登录xlogin网页获得token，然后访问类似（域名替换成自己的） http://blog.example.com/xc/edit/m1 （注意要带上URL参数token=自己刚刚登录获得的token），即可编辑Markdown格式的文件内容，位置在服务器/mnt/xms/pages目录下的abd.md文件。编辑后保存。然后访问 http://blog.example.com/xc/md/abc 即可访问渲染后的网页，同理 http://blog.example.com/xc/t/t1 可访问纯文本格式的t1.txt文件， http://blog.example.com/xc/h/a1 可访问网页格式的a1.html文件。 http://blog.example.com/xc/editxms/abc.xie 则是编辑一个谢语言代码文件，该文件保存后位于/mnt/xms/x目录下，之后可以用 http://blog.example.com/xms/x/abc 来访问该服务。一个例子文件如下，

After running, first log in to the xlogin webpage to obtain the token, and then visit something similar (replace the domain name with your own) http://blog.example.com/xc/edit/m1 (Note that you need to bring the URL parameter token=the token you just logged in to obtain) to edit the content of the Markdown format file, located in the abd.md file in the server/mnt/xms/pages directory. Save after editing. Then visit http://blog.example.com/xc/md/abc You can access the rendered webpage, similarly http://blog.example.com/xc/t/t1 Accessible t1.txt file in plain text format, http://blog.example.com/xc/h/a1 Accessible a1.html file in web format. http://blog.example.com/xc/editxms/abc.xie It is to edit a Xielang code file, which is saved and located in the/mnt/xms/x directory. Afterwards, you can use the http://blog.example.com/xms/x/abc To access the service. An example file is as follows:,

```
= $outG "TX_END_RESPONSE_XT"

setRespHeader $responseG "Content-Type" "text/json; charset=utf-8"

writeRespHeader $responseG #i200

pl "[%v] %v params: %v" @'{nowStr}' $reqNameG $paraMapG

genResp $rs $requestG success test2

writeResp $responseG $rs

exit
```

还可以进一步扩展功能，但一个简单的博客系统或者叫CMS（内容管理系统）已经搭建完成了。

The function can be further extended, but a simple blog system or content management system (CMS) has been built.

&nbsp;

##### - **嵌套运行谢语言代码**（Nested Run Xielang Code）

&nbsp;

谢语言中也可以另起一个虚拟机执行一段谢语言代码（即嵌套执行），某些情况下，这会是个很方便的功能。示例如下（runCode.xie）：

In Xielang, another virtual machine started by a certain VM can also execute a section of Xielang code (i.e. nested execution), which can be a very convenient feature in some cases. An example is as follows (runCode.xie):

```go
// 设定传入参数inputT，在虚拟机中通过全局变量inputG访问
// Set the input parameter inputT and access it in the virtual machine through the global variable inputG
assign $inputT #L`[{"name": "tom", "age": 25}, 15]`

// 用runCode指令运行代码
// 代码将在新的虚拟机中执行
// 除结果参数（不可省略）外，第一个参数是字符串类型的代码（必选，后面参数都是可选）
// 第二个参数为任意类型的传入虚拟机的参数（虚拟机内通过inputG全局变量来获取该参数）
// 第三个参数可以是一个列表，键值对将依次传入新虚拟机作为全局变量，这两个参数（第二、三个）如果不需要可以传入$nilG
// 再后面的参数可以是一个字符串数组类型的变量或者多个字符串类型的变量，虚拟机内通过argsG（字符串数组）来对其进行访问
// Running code with the runCode instruction
// The code will be executed in the new virtual machine
// Except for the result parameter (which cannot be omitted), the first parameter is the code of string type (required, all subsequent parameters are optional)
// The second parameter is any type of parameter passed into the virtual machine (obtained through the inputG global variable within the virtual machine)
// The third parameter can be a list, where key value pairs will be passed to the new virtual machine as global variables in sequence. If these two parameters (second and third) are not needed, $nilG can be passed in
// The following parameters can be a variable of string array type or multiple variables of string type, which are accessed through argsG (string array) in the virtual machine
runCode $result `

// 输出inputG供参考
// Output inputG for reference
pln "inputG=" $inputG

// 获取inputG中的第二项（序号为1，值为数字15）
// Obtain the second item in inputG (sequence number 1, value 15)
getItem $item2 $inputG 1

plo $item2

// 由于数字可能被JSON解析为浮点数，因此将其转换为整数
// Since the number may be parsed as a floating point number by JSON, it is converted to an integer
toInt $item2 $item2

// 输出argsG供参考
// Output argsG for reference
pln "argsG=" $argsG

// 从argsG中获取第一项（序号为0）
// Get the first item from argsG (sequence number 0)
getItem $v3 $argsG 0

// 由于argsG中每一项都是字符串，因此将其转换为整数
// Since each item in argsG is a string, convert it to an integer
toInt $v3 $v3

// 从argsG中获取第二项（序号为1）
// Get the second item (sequence number 1) from argsG
getItem $v4 $argsG 1

toInt $v4 $v4

// 定义一个变量a并赋值为整数6
// Define a variable a and assign it as an integer 6
assign $a #i6

// 用eval指令计算几个数相加的值，结果入栈
// 由于虚拟机已经用了反引号括起代码
// 因此可以用双引号括起表达式以免冲突
// Calculate the value of adding several numbers using the eval instruction, and push the result onto the stack
// Because the virtual machine has already used back quotes to enclose the code
// Therefore, expressions can be enclosed in double quotation marks to avoid conflicts
eval "$a + $item2 + $v3 + $v4"

// 设置虚拟机的返回值
// Set the return value of the virtual machine
assign $outG $tmp

` $inputT $nilG 22 9

// 最后结果应为52
// The final result should be 52
pln result= $result

```

需要注意的是输入参数和输出参数的运用方法，以及在嵌入代码中尽量避免使用反引号，如果确实会用到，可以用字符串替换的方法来解决。

It should be noted that the application methods of input and output parameters, as well as avoiding the use of backquotes in embedded code, can be solved by replacing strings if necessary.

运行结果如下：

The operation results are as follows:

```shell
inputG= [map[age:25 name:tom] 15]
(float64)15
argsG= [22 9]
result= 52
```

&nbsp;

#### **谢语言做系统服务**（Write system services by Xielang）

谢语言可以作为系统服务启动，支持Windows和Linux等操作系统。只要加命令行参数-reinstallService运行谢语言主程序，即可在系统中安装一个名为xieService的系统服务（在Windows下可以用计算机管理中的服务管理模块看到）。注意，操作系统中安装服务，一般需要管理员权限才可以进行，Windows下需要以管理员身份打开CMD窗口执行该命令，Linux下需要以root用户或用sudo命令来执行。

Xielang can be started as a system service and supports operating systems such as Windows and Linux. As long as you add the command line parameter '-reinstallService' to run the Xielang main program, you can install a system service called xieService in the system (which can be seen using the service management module in computer management under Windows). Note that installing services in the operating system generally requires administrator privileges. Under Windows, you need to open the CMD window as an administrator to execute this command, while under Linux, you need to execute it as root or with the sudo command.

服务启动后会在服务根目录（Windows下为c:\xie，Linux下为/xie）下的文件中xieService.log记录日志。服务初次启动时，会在服务根目录下寻找所有名称类似taskXXX.xie的文件（例如task001.xie、taskAbc.xie等）逐个运行，并将其执行结果（通过全局变量outG返回）输出到日志。这种代码文件称为一次性运行任务文件，一般用于需要开机运行一次的情况，也可以通过手动执行xie -restartService命令来重启服务达到再次执行的目的。

After the service is started, a log will be recorded in the file xieService.log in the service root directory (c:\xie in Windows and /xie in Linux). When the service starts for the first time, it will search for all files with names similar to taskXXX. xie in the service root directory (such as task001.xie, taskAbc.xie, etc.) and run them one by one, and output their execution results (returned through the global variable outG) to the log. This type of code file is called a one-time run task file, and is generally used in situations where it needs to be started and run once. It can also be manually run the command 'xie -restartService' to restart the service and achieve the goal of re-execution.

另外，xieService服务在运行中，每隔5秒钟会检查服务根目录，如果其中有名称类似autoRemoveTaskXXX.xie的文件（例如autoRemoveTask001.xie、autoRemoveTaskAbc.xie等），将会立即执行这些文件中的代码，然后将这些文件删除。这种机制类似任务队列，允许我们随时将任务加入队列（放入服务根目录），谢语言服务会随时执行这些任务。并且由于执行后会立即删除，因此该任务不会被反复执行。

In addition, during operation, the xieService service checks the service root directory every 5 seconds. If there are files with names similar to autoRemoveTaskXXX.xie (such as autoRemoveTask001.xie, autoRemoveTaskAbc.xie, etc.), the code in these files will be immediately executed and then deleted. This mechanism is similar to a task queue, allowing us to add tasks to the queue (placed in the service root directory) at any time, and Xielang service will execute these tasks at any time. And since the task will be deleted immediately after execution, it will not be executed repeatedly.

与服务安装、移除、启动、停止、重新启动有关的谢语言主程序命令行参数还包括-installService、-removeService、-startService、-stopService、-restartService等。

The command line parameters related to service installation, removal, start, stop, and restart of the Xielang main program also include '-installService', '-removeService', '-startService', '-stopService', '-restartService', and so on.

任务代码可以参考例子中的task001.xie、autoRemoveTask001.xie等。

The task code can refer to examples such as task001.xie, autoRemoveTask001.xie, etc.

&nbsp;

#### **图形界面（GUI）编程**（GUI Programming）

谢语言支持方便的图形界面（GUI）编程，包含多种实现方式，各有各的优势和使用场景。

Xielang supports convenient graphical interface (GUI) programming, including multiple implementation methods, each with its own advantages and usage scenarios.

其中，Windows下使用WebView2系统控件是比较推荐的GUI编程方式，WebView2功能强大并且随时更新，在Windows 10及以上系统中已经内置，Windows 7等系统中也可以单独安装，谢语言无需附加任何文件即可用这种方式编写和分发图形界面应用。

Among them, using WebView2 system controls under Windows is a recommended GUI programming method. WebView2 is powerful and constantly updated, and is already built-in in Windows 10 and above systems. It can also be installed separately in Windows 7 and other systems. Xielang can write and distribute graphical interface applications in this way without attaching any files.

另一种方式是使用一个外部的浏览器来访问谢语言启动的WEB服务器或API服务器，这样前端可以完全使用标准的HTML/CSS/JavaScript技术进行图形界面编程，通过Ajax方式访问谢语言编写的Web服务来使用谢语言的能力。这种方式的缺点是，一般的浏览器为安全考虑一般不允许通过代码调整浏览器的标题、大小和位置。

The other way is to use an external browser to access the web server or API server launched by Xielang, so that the front-end can fully use standard HTML/CSS/JavaScript technology for graphical interface programming, and access the web services written in Xielang through Ajax to use Xielang's capabilities. The disadvantage of this approach is that general browsers do not allow code to adjust the title, size, and position of the browser for security reasons.

还有种方式是在第二种方式的基础上，调用谢语言配套的浏览器，即可解决调整浏览器标题、大小和位置等问题。目前谢语言配套的浏览器包括一个基于WebView2控件的浏览器，可在Windows 7以上的系统中运行（Windows 7中需要单独安装WebView2控件）。

Another method is based on the 2nd method, calling the browser supporting Xielang can solve the problem of adjusting the browser title, size, and position. At present, the browsers supported by Xielang include a browser based on the WebView2 control. It can run on systems from Windows 7 and later.

谢语言中的图形界面编程通过下面的基本说明和几个例子可以快速地了解掌握。

The graphical interface programming in Xielang can be quickly understood and mastered through the following basic explanations and a few examples.

&nbsp;

#### **谢语言GUI编程的基础（WebView2）** （Fundamentals of GUI Programming in Xielang (WebView2)）

谢语言GUI图形编程的WebView2方式，主要通过Windows自带的WebView2组件来支持GUI编程，仅适用于Windows系统，分发时无需附加文件（如果低版本Windows系统，可以自行下载安装WebView2）。WebView2使用标准的HTML、CSS以及JavaScript的进行编程，来实现图形界面的展示和操控，谢语言则负责后台逻辑的处理，两者之间可以互通，JavaScript中通过特定的接口方式可以调用谢语言中的函数传递数据并进行操作，反之亦然，谢语言也可以调用JavaScript中的特定函数。基本熟悉网页编程的开发者都可以很方便地上手。

The WebView2 method of GUI graphical programming in Xielang mainly supports GUI programming through the built-in WebView2 component of Windows. It is only applicable to Windows systems and does not require additional files for distribution (if you have a lower version of Windows system, you can download and install WebView2 yourself). WebView2 uses standard HTML, CSS, and JavaScript programming to display and manipulate graphical interfaces, while Xielang is responsible for processing backend logic. The two can communicate with each other. JavaScript can call functions in Xielang to transmit data and perform operations through specific interface methods, and vice versa. Xielang can also call specific functions in JavaScript. Developers who are basically familiar with web programming can easily get started.

谢语言中有一个预置全局变量\$guiG，用于作为调用GUI功能的接口对象。

In Xielang, there is a preset global variable  $guiG used as an interface object for calling GUI functions.

下面我们通过一些例子逐步说明谢语言中基于WebView2方式的GUI编程方法。

Below, we will gradually illustrate the GUI programming method based on WebView2 in Xielang through some examples.

&nbsp;

##### - **基本界面**（Basic GUI）

我们直接通过一个代码例子（webGui1.xie）来了解：

We can directly understand through a code example (webGui1.xie):

```go
// 本例演示使用Windows下的WebView2（Windows 10以上自带，Win 7等可以单独安装）来制作图形化界面程序
// WebView2在Windows 10以上系统自带，Win 7等可以单独安装
// 也因此本例只在Windows下有效
// This example demonstrates using WebView2 under Windows (which comes with Windows 10 or above, and can be installed separately for Win 7 or other applications) to create a graphical interface program
// WebView2 comes with Windows 10 and above systems, and Win 7 and others can be installed separately
// Therefore, this example is only valid under Windows

// 新建一个窗口，放入变量w中
// guiG是全局预置变量，表示图形界面主控对象
// 它的newWindow方法根据指定参数创建一个新窗口
// width参数表示窗口的宽度，缺省为800
// height参数表示窗口的高度，缺省为600
// 如果带有-debug参数，表示是否允许调试（鼠标右键菜单带有“检查”等选项）
// -fix参数表示窗口不允许调整大小
// -center参数表示窗口居中
// 还有-max、-min分别表示以最大或最小化的状态展示窗口
// Create a new window and place it in the variable w
// guiG is a global preset variable that represents the main control object of the graphical interface
// Its newWindow method creates a new window based on the specified parameters
// The width parameter represents the width of the window, which defaults to 800
// The height parameter represents the height of the window, which defaults to 600
// If there is a -debug parameter, it indicates whether debugging is allowed (the right-click menu has options such as "check")
// The -fix parameter indicates that the window does not allow resizing
// The -center parameter indicates that the window is centered
// Also, -max and -min represent displaying windows in maximum or minimum states, respectively
mt $w $guiG newWindow "-title=Test WebView2" -width=1024 -height=768 -center

plo $w

// 用于网页中的快速代理函数代码
// 网页中的JavaScript代码中可以用quickDelegateDo函数来调用本函数
// 快速代理函数将在新的运行上下文中执行
// quickDelegateDo函数中所带的参数将被封装成一个列表（数组）放入$inputL变量中
// 快速代理函数中可以对其按索引取值进行处理
// 快速处理函数也可以使用虚拟机级的全局变量、寄存器或堆栈进行数据共享
// Quick proxy function code for web pages
// The quickDelegateDo function can be used in JavaScript code on web pages to call this function
// The fast proxy function will be executed in the new runtime context
// The parameters carried in the quickDelegateDo function will be encapsulated into a list (array) and placed in the $inputL variable
// Fast proxy functions can be processed based on index values
// Fast processing functions can also use virtual machine level global variables, registers, or stacks for data sharing
= $dele1 `
	// 输出变量inputL供参考
	// Output variable inputL for reference
    pl "%#v" $inputL
    
	// 本例中，第一个参数被约定为传递一个命令
	// 后面的参数为该命令所需的参数，参数个数视该命令的需要而定
	// 因此这里从参数数组中取出第一个参数放入变量cmdT中
	// In this example, the first parameter is specified to pass a command
	// The following parameters are required for the command, and the number of parameters depends on the needs of the command
	// Therefore, the first parameter is taken from the parameter array and placed in the variable cmdT here
    getArrayItem $cmdT $inputL 0

	// 如果命令为showNav，则取后两个参数并输出其内容
	// If the command is showNav, take the last two parameters and output their contents
    ifEqual $cmdT "showNav" :+1 :inext1
        getArrayItem $arg1 $inputL 1
        getArrayItem $arg2 $inputL 2

        pl "name: %v, value: %v" $arg1 $arg2

		// 快速处理函数最后必须通过变量outL返回一个值，无论是否需要
		// The fast processing function must ultimately return a value through the variable outL, regardless of whether it needs to be
        = $outL "showNav result"

		// 快速处理函数最后用exit指令返回
		// Quickly process the function and return it with the exit instruction
        exitL

    :inext1
	// 如果命令为pl，则类似pl指令（其他语言中的或printf）
	// 取出后面第一个参数为格式化字串
	// 再后面都是格式化字串中所需的填充值
	// 然后输出到标准输出
	// If the command is pl, it is similar to the pl instruction (in other languages as printf with an extra line-end)
	// Take out the first parameter that follows as a formatted string
	// The following are the required padding values in the formatted string
	// Then output to standard output
    ifEqual $cmdT "pl" :+1 :inext2
        getArrayItem $formatT $inputL 1

		// 截取inputL中第三项（序号为2）开始的所有项
		// Get all items starting from the third item (sequence number 2) in inputL
        slice $list1 $inputL 2 -

		// 用pl指令输出指定的内容，注意“$list1...”写法表示展开其中的列表参数
		// Use the pl instruction to output the specified content. Note that the notation '$list1...' indicates expanding the list parameters within it
        pl $formatT $list1...

		// 注意exitL指令后可以跟随一个参数，该参数将自动被放入$outL中，这是一种简化的函数返回的写法
		// Note that the exitL instruction can be followed by a parameter that will automatically be placed in $outL, which is a simplified method of writing function returns
        exitL "exit from pl"

    :inext2
	// 不支持的命令将输出错误信息
	// Output error messages for unsupported commands
    pl "unknown command: %v" $cmdT

    exitL @'{spr $tmp "unknown command: %v" $cmdT}'
`

// 新建一个用于窗口事件处理的快速代理函数
// 代码存于变量$dele1中
// 快速代理函数必须以exitL指令返回
// Create a new fast proxy function for window event processing
// Code stored in variable $dele1
// The fast proxy function must return with the exitL instruction
new $deleT quickDelegate $dele1

checkErrX $deleT

// 调用窗口对象的setQuickDelegate方法来指定代理函数
// Call the setQuickDelegate method of the window object to specify the proxy function
mt $rs $w setQuickDelegate $deleT

plo $rs

// 如果从网络加载网页，那么可以用下面的navigate方法
// mt $rs $w navigate http://xie.topget.org
// If you load a webpage from the network, you can use the navigate method below
// mt $rs $w navigate http://xie.topget.org

// 本例中使用从本地加载的网页代码
// 设置准备在窗口中载入的HTML代码
// 本例中HTML页面中引入的JavaScript和CSS代码均直接用网址形式加载
// In this example, the webpage code loaded locally is used
// Set the HTML code to be loaded in the window
// In this example, the JavaScript and CSS code introduced in the HTML page are directly loaded in the form of website addresses
= $htmlT `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="content-type" content="text/html; charset=UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<script src="http://xie.topget.org/js/jquery.min.js"></script>
<link rel="stylesheet"  type="text/css" href="http://xie.topget.org/css/tabulator.min.css">
<script src="http://xie.topget.org/js/tabulator.min.js"></script>
<title></title>
<script>
	// 页面加载完毕后，将用alert展示一个值，然后准备数据并显示一个报表
	// After the page is loaded, an alert will be used to display a value, and then the data will be prepared and a report will be displayed
	window.onload = function() {
		var s1 = "a信b";

		var s2 = "1\x602";
		alert(s2);

		console.log(s1.charCodeAt(0), s1.charCodeAt(1), s1.charCodeAt(2), s2, JSON.stringify(s2));

		var tabledata = [
            {id:1, name:"Oli Bob", age:"12", col:"red", dob:""},
            {id:2, name:"Mary May", age:"1", col:"blue", dob:"14/05/1982"},
            {id:3, name:"Christine Lobowski", age:"42", col:"green", dob:"22/05/1982"},
            {id:4, name:"Brendon Philips", age:"125", col:"orange", dob:"01/08/1980"},
            {id:5, name:"Margret Marmajuke", age:"16", col:"yellow", dob:"31/01/1999"},
        ];

		var table = new Tabulator("#div3", {
			height:205, // set height of table (in CSS or here), this enables the Virtual DOM and improves render speed dramatically (can be any valid css height value)
			data:tabledata, //assign data to table
			layout:"fitColumns", //fit columns to width of table (optional)
			columns:[ //Define Table Columns
				{title:"Name", field:"name", width:150},
				{title:"Age", field:"age", hozAlign:"left", formatter:"progress"},
				{title:"Favourite Color", field:"col"},
				{title:"Date Of Birth", field:"dob", sorter:"date", hozAlign:"center"},
			],
			rowClick:function(e, row){ //trigger an alert message when the row is clicked
				alert("Row " + row.getData().id + " Clicked!!!!");
			},
		});

	}

	// 点击test1按钮后，将调用quickDelegateDo函数来调用谢语言中定义的快速代理函数，并传入需要的参数，然后alert返回的值
	// After clicking the test1 button, the quickDelegateDo function will be called to call the fast proxy function defined in Xielang, passing in the required parameters, and then the value returned by alert
	function test1() {
		var rs = quickDelegateDo("pl", "time: %v, navigator: %v", new Date(), navigator.userAgent);

		// 返回的结果是一个Promise，因此要用相应的方式获取
		// The returned result is a Promise, so it needs to be obtained in a corresponding way
		rs.then(res => {
			alert("test1: "+res);
		});
	}

	// 点击test2按钮后，将调用quickDelegateDo函数来调用谢语言中定义的快速代理函数，并alert返回的值
	// After clicking the test2 button, the quickDelegateDo function will be called to call the fast proxy function defined in Xielang, and the returned value will be alerted
	function test2() {
		var rs = quickDelegateDo("showNav", "userAgent", navigator.userAgent);

		// 返回的结果是一个Promise，因此要用相应的方式获取
		// The returned result is a Promise, so it needs to be obtained in a corresponding way
		rs.then(res => {
			alert("test2: "+res);
		});
	}

	// 点击test按钮后，将用Ajax方式访问一个网络API，获取结果并显示
	// After clicking the test button, a network API will be accessed using Ajax to obtain the results and display them
	function test() {
		$.ajax({
			url: "http://topget.org/xms/test",
			dataType: 'text',
			type: 'POST',
			data: { 
				req: "test", 
				name: 'Tom'
			},
			success: function (data) {
				alert(data);
			},
			error: function (response) {
				console.log(JSON.stringify(response));
			}
		});

	}
</script>
</head>
<body>
<div id="div1">
	<button onclick="javascript:test();">test</button>
	<button onclick="javascript:test1();">test1</button>
	<button onclick="javascript:test2();">test2</button>
</div>
<div id="div3">
</div>
</body>
</html>
`

// 调用窗口对象的setHtml方法来设置其内容
// Call the setHtml method of the window object to set its content
mt $rs $w setHtml $htmlT

plo $rs

// 调用窗口对象的setHtml方法来展示窗口
// 此时窗口才真正显示
// 并且直至窗口关闭都将阻塞（即等待窗口关闭后才往下继续执行后面的代码）
// Call the setHtml method of the window object to display the window
// At this point, the window truly displays
// And it will block until the window closes (i.e. wait for the window to close before continuing to execute the following code)
mt $rs $w show

plo $rs

// 调用窗口对象的close方法关闭窗口
// Calling the close method of the window object to close the window
mt $rs $w close

plo $rs

// 结束程序的执行
// End program execution
exit


```

代码运行后，将看到类似下面的界面：

After running the code, you will see an interface similar to the following:

![截图/snapshot](http://xie.topget.org/example/xie/snap/snap8.png)

代码中有详尽注释，我们可以看到，代码中展示了如何载入一个HTML页面作为窗口并显示出来，点击几个test按钮可以进行不同的操作，其中test1和test2都是与谢语言的后台逻辑进行互动，其中test1、test2还从谢语言处理函数中获取了返回值并显示。test按钮则演示了如何通过Ajax方式获取一个网络API请求的结果并进行处理。

There are detailed annotations in the code, which shows how to load an HTML page as a window and display it. Clicking a few test buttons can perform different operations. Among them, test1 and test2 interact with the backend logic of Xielang, and test1 and test2 also obtain return values from Xielang processing functions and display them. The test button demonstrates how to obtain the result of a network API request through Ajax and process it.

&nbsp;

##### - **直接嵌入网页脚本**（Directly embed JavaScript file in WEB pages）

下面的这个代码例子（webGui2.xie）与上面类似，但使用了内置嵌入JavaScript或CSS文本的方式，避免了网络访问或者从附带文件中读取的麻烦。另外，本例中还演示了如何设置更安全的代理（回调）函数来进行前台界面与谢语言后台的互动。

The following code example (webGui2.xie) is similar to the above, but uses built-in embedded JavaScript or CSS text to avoid the hassle of network access or reading from accompanying files. In addition, this example also demonstrates how to set up a more secure proxy (callback) function to interact with the front-end interface and Xielang backend.

```go
// 本例演示使用WebView2做图形界面时
// 获取内置的JavaScript或CSS文本嵌入HTML中
// 这样可以避免网络访问或者从附带文件中读取的麻烦
// 另外，本例也演示了如何设置普通代理函数来更安全地进行网页与谢语言后台逻辑之间的互动
// This example demonstrates using WebView2 as a graphical interface
// Get built-in JavaScript or CSS text embedded in HTML
// This can avoid the hassle of network access or reading from accompanying files
// In addition, this example also demonstrates how to set a regular proxy function to more securely interact between web pages and Xielang backend logic

// guiNewWindow是内置指令，与下面命令等效
// guiNewWindow is a built-in instruction that is equivalent to the following commands
// mt $w $guiG newWindow "-title=Test WebView2a" -width=1024 -height=768 -center -debug
// -debug参数表示打开调试功能
// The -debug parameter indicates that debugging is enabled
guiNewWindow $w "-title=Test WebView2a" -width=1024 -height=768 -center -debug

// 如果出错则停止执行
// Stop execution if an error occurs
checkErrX $w

// 调用窗口对象的setDelegate方法来指定代理函数
// 之前的例子中使用的快速代理函数直接在当前虚拟机中运行，存在一定的并发冲突可能性
// 因此为安全起见，更建议使用普通代理函数
// 普通代理函数通过字符串来定义其代码
// 普通代理函数将在单独新建的虚拟机中运行
// 传入的参数通过全局变量inputG传入，是一个参数数组
// 传出的参数则应放于全局outG中返回
// 与快速代理函数不同，普通代理函数不用exitL指令来退出，而是直接用exit指令
// Call the setDelegate method of the window object to specify the proxy(callback) function
// The fast proxy function used in the previous example runs directly on the current virtual machine, which has a certain possibility of concurrency conflicts
// Therefore, for safety reasons, it is more recommended to use regular proxy functions
// A regular proxy function defines its code through a string
// The regular proxy function will run in a newly created virtual machine separately
// The passed in parameters are passed in through the global variable inputG, which is an array of parameters
// The outgoing parameters should be placed in the global outG and returned
// Unlike fast proxy functions, regular proxy functions do not use the exitL instruction to exit, but instead use the exit instruction directly
mt $rs $w setDelegate `
     
    getArrayItem $cmdT $inputG 0

    ifEqual $cmdT "showNav" :+1 :inext1
        getArrayItem $arg1 $inputG 1
        getArrayItem $arg2 $inputG 2

        pl "name: %v, value: %v" $arg1 $arg2

        = $outG "showNav result"

        exit

    :inext1
    ifEqual $cmdT "pl" :+1 :inext2
        getArrayItem $formatT $inputG 1

        slice $list1 $inputG 2 -

        pl $formatT $list1...

        = $outG ""

        exit

    :inext2
    pl "unknown command: %v" $cmdT

    spr $outG "unknown command: %v" $cmdT

    exit
`

= $htmlT `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="content-type" content="text/html; charset=UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<script>TX_jquery.min.js_XT</script>
<style>TX_tabulator.min.css_XT</style>
<script>TX_tabulator.min.js_XT</script>
<title></title>
<script>
	$().ready(function (){
		var tabledata = [
            {id:1, name:"Oli Bob", age:"12", col:"red", dob:""},
            {id:2, name:"Mary May", age:"1", col:"blue", dob:"14/05/1982"},
            {id:3, name:"Christine Lobowski", age:"42", col:"green", dob:"22/05/1982"},
            {id:4, name:"Brendon Philips", age:"125", col:"orange", dob:"01/08/1980"},
            {id:5, name:"Margret Marmajuke", age:"16", col:"yellow", dob:"31/01/1999"},
        ];

		var table = new Tabulator("#div3", {
			height:205, // set height of table (in CSS or here), this enables the Virtual DOM and improves render speed dramatically (can be any valid css height value)
			data:tabledata, //assign data to table
			layout:"fitColumns", //fit columns to width of table (optional)
			columns:[ //Define Table Columns
				{title:"Name", field:"name", width:150},
				{title:"Age", field:"age", hozAlign:"left", formatter:"progress"},
				{title:"Favourite Color", field:"col"},
				{title:"Date Of Birth", field:"dob", sorter:"date", hozAlign:"center"},
			],
			rowClick:function(e, row){ //trigger an alert message when the row is clicked
				alert("Row " + row.getData().id + " Clicked!!!!");
			},
		});

	});

	function test1() {
		delegateDo("pl", "time: %v, navigator: %v", new Date(), navigator.userAgent);
	}

	function test2() {
		var rs = delegateDo("showNav", "userAgent", navigator.userAgent);

		rs.then(res => {
			alert("test2: "+res);
		});
	}
</script>
</head>
<body>
<div id="div1">
	<button onclick="javascript:test1();">test1</button>
	<button onclick="javascript:test2();">test2</button>
</div>
<div id="div3" style="margin-top: 1.0em;">
</div>
</body>
</html>
`

// 提示：使用getResourceList指令可以看到所有内置的资源
// Tip: Use the getResourceList directive to see all built-in resources
getResource $t1 "js/jquery.min.js"

strReplace $htmlT $htmlT "TX_jquery.min.js_XT" $t1

getResource $t2 "css/tabulator.min.css"

strReplace $htmlT $htmlT "TX_tabulator.min.css_XT" $t2

getResource $t3 "js/tabulator.min.js"

strReplace $htmlT $htmlT "TX_tabulator.min.js_XT" $t3

mt $rs $w setHtml $htmlT

checkErrX $rs

mt $rs $w show

checkErrX $rs

mt $rs $w close

exit

```

&nbsp;

##### - **启动后台服务与前台配合**（Start the backend service and cooperate with the front end）

下面的这个代码例子（webGui3.xie）与前两个也是类似，但前后台并非采用回调函数来进行互动，而是谢语言后台用线程在本机的随机端口上启动了一个WEB与API混合服务器来提供网页与接口服务，前台WebView2通过HTTP协议来访问后台接口实现互动，这也是常见的一种方式。

The following code example (webGui3.xie) is also similar to the previous two, but the front-end and back-end do not use callback functions for interaction. Instead, the Xielang back-end uses a thread to start a web and API mixed server on a random port on the local machine to provide web page and interface services. The front-end WebView2 accesses the back-end interface through the HTTP protocol to achieve interaction, which is also a common way.

```go
// 本例演示使用WebView2做图形界面时
// 启动一个谢语言WEB服务器和API服务器来自行提供网页资源与API数据服务
// 这样可以避免网络访问或者从附带文件中读取的麻烦，实现前后台的互通
// 唯一的缺点是需要占用一个本机端口
// This example demonstrates using WebView2 as a graphical interface
// Start a Xielang web server and API server to provide web resources and API data services on your own
// This can avoid the trouble of network access or reading from accompanying files, and achieve interoperability between the front and back ends
// The only drawback is that it requires occupying a local port

guiNewWindow $w "-title=Test WebView2b" -width=1024 -height=768 -center -debug

checkErrX $w

// 设置路由处理器
// Set Routing Processor
newMux $muxT

// 设置静态内容的处理函数
// 用于网页中嵌入JS和CSS时获取内置资源中的这些内容
// 这样，如果主页的网址是 http://127.0.0.1:8721
// 那么，网页中可以用嵌入的 /static/js/jquery.min.js 来获取内置的内容
// Set the processing function for static content
// Used to obtain these contents from built-in resources when embedding JS and CSS in web pages
// So, if the website address of the homepage is http://127.0.0.1:8721
// So, embedded/static/js/jquery.min.js can be used in web pages to obtain built-in content
setMuxHandler $muxT "/static/" "" `
	// 去掉请求路由的前缀 /static/
	// Remove prefix/static from request routing/
	trimPrefix $shortNameT $reqNameG "/static/"

	// 获取形如 js/jquery.min.js 形式的内置资源内容
	// Obtain built-in resource content in the form of js/jquery.min.js
	getResource $textT $shortNameT

	// 根据内置资源的后缀名，获取其MIME类型，例如：text/javascript
	// Obtain the MIME type of the built-in resource based on its suffix name, for example: text/JavaScript
	getMimeType $mimeTypeT $shortNameT

	// 拼装完整的mime类型字符串
	// Assemble complete mime type strings
	spr $mimeTypeT "%v; charset=utf-8" $mimeTypeT 

	setRespHeader $responseG "Content-Type" $mimeTypeT
	writeRespHeader $responseG 200

	writeResp $responseG $textT

	assign $outG "TX_END_RESPONSE_XT"

`

// 设置/test路由处理函数，用于测试WEB API
// 返回内容是JSON格式
// Set '/test' routing processing function for testing WEB API
// The returned content is in JSON format
setMuxHandler $muxT "/test" 0 `
	setRespHeader $responseG "Content-Type" "text/json; charset=utf-8"
	writeRespHeader $responseG 200

	spr $strT "[%v] Req: test，Parameters： %v，inputG：%v" @'{nowStr}' $paraMapG $inputG

	var $resMapT map

	setMapItem $resMapT "Status" "success"
	setMapItem $resMapT "Value" $strT

	toJson $jsonStrT $resMapT

	writeResp $responseG $jsonStrT

	assign $outG  "TX_END_RESPONSE_XT"
`

// htmlT中即为准备用于根路由访问时的网页
// 其中 test、test1和test2函数分别演示了使用异步Ajax、fetch和同步Ajax方式来调用本地接口的例子
// The webpage in HTMLT is prepared for root routing access
// The test, test1, and test2 functions demonstrate examples of using asynchronous Ajax, fetch, and synchronous Ajax methods to call local interfaces, respectively
= $htmlT `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="content-type" content="text/html; charset=UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<script src="/static/js/jquery.min.js"></script>
<link rel="stylesheet" type="text/css" href="/static/css/tabulator.min.css">
<script src="/static/js/tabulator.min.js"></script>
<script>
	$().ready(function (){
		var tabledata = [
            {id:1, name:"Oli Bob", age:"12", col:"red", dob:""},
            {id:2, name:"Mary May", age:"1", col:"blue", dob:"14/05/1982"},
            {id:3, name:"Christine Lobowski", age:"42", col:"green", dob:"22/05/1982"},
            {id:4, name:"Brendon Philips", age:"125", col:"orange", dob:"01/08/1980"},
            {id:5, name:"Margret Marmajuke", age:"16", col:"yellow", dob:"31/01/1999"},
        ];

		var table = new Tabulator("#div3", {
			height:205,
			data:tabledata, 
			layout:"fitColumns", 
			columns:[ 
				{title:"Name", field:"name", width:150},
				{title:"Age", field:"age", hozAlign:"left", formatter:"progress"},
				{title:"Favourite Color", field:"col"},
				{title:"Date Of Birth", field:"dob", sorter:"date", hozAlign:"center"},
			],
			rowClick:function(e, row){ 
				alert("Row " + row.getData().id + " Clicked!!!!");
			},
		});

	});

	function test1() {
		fetch('/test', {
			method: 'POST', 
			body: JSON.stringify({
				time: new Date(),
				navigator: navigator.userAgent
			})
		}).then(function(res) { 
			res.json().then(function(res1){
				alert(JSON.stringify(res1));
			});
		});
	}

	function test2() {
		var rs = $.ajax({
			url: "/test",
			type: "POST",
			async: false,
			dataType: "text",
			data: {
				req: "test", 
				name: 'Jerry'
			}
		});

		var objT = JSON.parse(rs.responseText);

		if (objT.Status == "success") {
			alert("success: " + objT.Value);
		} else {
			alert("fail: " + objT.Value);
		}
	}

	function test() {
		$.ajax({
			url: "/test",
			dataType: 'text',
			type: 'POST',
			data: { 
				req: "test", 
				name: 'Tom'
			},
			success: function (data) {
				alert(data);
			},
			error: function (response) {
				console.log(JSON.stringify(response));
			}
		});

	}

</script>
</head>
<body>
<div id="div1">
	<button onclick="javascript:test();">test</button>
	<button onclick="javascript:test1();">test1</button>
	<button onclick="javascript:test2();">test2</button>
</div>
<div id="div3" style="margin-top: 1.0em;">
</div>
</body>
</html>
`

// 设置根路径访问时的返回内容
// 即htmlT中存放的网页HTML
// setMuxHandler中的第三个参数传入处理函数中即为可通过全局变量inputG访问的值
// Set the return content when accessing the root path
// The webpage HTML stored in htmlT
// The third parameter in setMuxHandler passed into the processing function is the value that can be accessed through the global variable inputG
setMuxHandler $muxT "/" $htmlT `
	setRespHeader $responseG "Content-Type" "text/html; charset=utf-8"
	writeRespHeader $responseG 200

	writeResp $responseG $inputG

	assign $outG "TX_END_RESPONSE_XT"
`


// 获取一个随机的可用端口用于命令服务器与图形界面通信
// Obtain a random available port for command server and graphical interface communication
getRandomPort $portT

// 启动一个线程来运行HTTP服务器
// Start a thread to run the HTTP server
startHttpServer $resultT $portT $muxT -go

spr $urlT "http://127.0.0.1:%v" $portT

// 让WebView2窗口访问本机的这个端口
// URL地址类似http://127.0.0.1:8721
// Enable the WebView2 window to access this port on the local machine
// URL address is similar http://127.0.0.1:8721
mt $rs $w navigate $urlT

checkErrX $rs

mt $rs $w show

checkErrX $rs

mt $rs $w close

exit

```

&nbsp;

##### - **简单的图形计算器**（A Simple GUI Calculator）

我们直接通过一个代码例子（calculatorGui.xie）来了解：

We can directly understand through a code example (calculatorGui.xie):

```go
// 定义用于界面展示的HTML网页代码，放在htmlT变量中
// HTML和CSS代码都是标准的，脚本语言也是标准的JavaScript
// 本例中定义了一个文本输入框用于输入表达式算式
// 以及“Calculate”和“Close”两个按钮
// 并定义了两个按钮对应的处理脚本函数
// “Calculate”按钮将调用JavaScript的eval函数来进行表达式计算
// 然后将计算结果传递给谢语言代码（通过调用谢语言预定义的quickDelegateDo函数）
// “Close”按钮将关闭整个窗口
// Define HTML web page code for interface display, placed in the htmlT variable
// HTML and CSS code are both standard, and the scripting language is also standard JavaScript
// In this example, a text input box is defined for inputting expression expressions
// And the "Calculate" and "Close" buttons
// And defined the processing script functions corresponding to the two buttons
// The 'Calculate' button will call JavaScript's eval function for expression evaluation
// Then pass the calculation results to Xielang code (by calling Xielang's predefined quickDelegateDo function)
//The 'Close' button will close the entire window
assign $htmlT `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <title>Calculator</title>
</head>
<body>
	<div style="margin-top: 10px; margin-bottom: 10px;">
		<span>Please enter the expression:</span>
	</div>
	<div style="margin-top: 10px; margin-bottom: 10px;">
		<input id="mainInputID" type=text />
	</div>
	<div>
		<button id="btnCal">Calculate</button>
		<button id="btnClose" onclick="javascript:closeWindowClick();">Close</button>
	</div>

    <script>
        document.getElementById("btnCal").addEventListener("click", function() {
			var result = eval(document.getElementById("mainInputID").value);

            quickDelegateDo(result);

            document.getElementById("mainInputID").value = result;
        });

        function closeWindowClick() {
            quickDelegateDo("closeWindow");
        }
 
        window.onload = function() {
        };
 
    </script>
</body>
</html>
`

// 调用guiG的newWindow方法创建一个窗口
// newWindow方法需要有三个参数，第一个是窗口标题
// 第二个是字符串形式的值用于指定窗口大小，空字符串表示按默认区域
// 如果使用类似“[200,300,600,400]”的字符串，则表明窗口位于屏幕坐标（200,300）处，宽高为600*400
// 第三个参数为用于界面展示的字符串
// 结果放入变量windowT中，这是一个特殊类型的对象(后面暂称为window对象)
// 后面我们还将调用该对象的一些方法进行进一步的界面控制
// Calling guiG's newWindow method to create a window
// The newWindow method requires three parameters, the first being the window title
// The second value is in string form to specify the window size, and an empty string represents the default area
// If a string similar to '[200,300,600,400]' is used, it indicates that the window is located at the screen coordinate (200,300), with a width of 600 * 400 in the high order
// The third parameter is the string used for interface display
// The result is placed in the variable windowT, which is a special type of object (later temporarily referred to as a window object)
// In the future, we will also call some methods of this object for further interface control
mt $windowT $guiG newWindow "-title=Simple Calculator" -width=640 -height=480 -center

plo $windowT

// 用new指令创建一个快速代理函数（quickDelegate）对象dele1
// 谢语言中quickDelegate是最常用的代理函数对象
// 它创建时需要指定一个快速函数，本例中通过源代码指明
// 这样，当网页代码中调用view对象的quickDelegateDo函数时
// 就将调用该快速函数代码来处理
// 约定该函数必须通过inputL变量来获取输入参数，并返回一个参数（通过outL变量）
// 参数均为字符串类型
// 如果传递复杂数据，常见的方法是传递JSON字符串
// 此处该函数仅仅是将输入参数输出
// Create a quickDelegate object dele1 using the new instruction
// In Xielang, quickDelegate is the most commonly used proxy function object
// When it is created, a fast function needs to be specified. In this example, the source code indicates
// In this way, when the web page calls the quickDelegateDo function of the view object
// Call the fast function code to handle
// The convention is that the function must obtain input parameters through the inputL variable and return a parameter (through the outL variable)
// All parameters are of string type
// If complex data is passed, a common method is to pass JSON strings
// This function only outputs the input parameters here
new $dele1 quickDelegate `
    [] $resultL $inputL 0

    ifEqual $resultL "closeWindow" :+1 :next1L
        mt $drop $windowT close
        exitL $resultL

    :next1L
    pl "Result: %v" $resultL

    // 函数返回前必须要有一个输出参数存入outL中
    // 此处因为实际上无需返回参数，因此随便存入一个无用的数值
    // There must be an output parameter stored in outL before the function returns
    // Because there is actually no need to return parameters, a useless numerical value is randomly stored here
    exitL $resultL
`

checkErrX $dele1

// 调用window对象的setDelegate方法将其接口代理指定为dele1
// Call the setDelegate method of the window object to specify its interface proxy as dele1
mt $rs $windowT setQuickDelegate $dele1

checkErrX $rs

mt $rs $windowT setHtml $htmlT

checkErrX $rs

// 调用window对象的show方法，此时才会真正显示界面窗口
// 并开始响应用户的操作
// Call the show method of the window object to truly display the interface window
// And start responding to user actions
mt $rs $windowT show

checkErrX $rs

mt $rs $windowT close

checkErrX $rs

// 退出程序
// Exit the program
exit


```

代码展示了如何用谢语言实现一个简单的图形界面计算器，代码中有详细的解释，可以仔细阅读理解。

代码运行后，将得到类似下面的界面：

After running the code, you will get an interface similar to the following:

![截图](http://xie.topget.org/example/xie/snap/snap9.png)

在输入框中输入算式，然后点击“Calculate”按钮，框中就会计算出结果，并且后台也得到了计算结果并将其输出。点击“关闭”按钮则窗口将关闭并执行后续代码（此例中是用exit指令退出了程序运行）。

Enter a formula in the input box, and then click the "Calculate" button. The result will be calculated in the box, and the calculation result will also be obtained in the background and output. Clicking the "Close" button will cause the window to close and execute subsequent code (in this case, the exit command was used to exit the program).

&nbsp;

##### - **Windows编译不带命令行窗口的谢语言主程序**（Compiling Xielang main program without command line window for Windows）

用谢语言在Windows系统下进行图形界面编程时，如果程序运行时不希望显示命令窗口（CMD），可以在编译谢语言源码（Go语言版）时加上-ldflags="-H windowsgui"的编译参数即可。

When using Xielang for graphical interface programming on Windows systems, if the program does not want to display a command window (CMD) during runtime, you can add the compilation parameter -ldflags="-H windowsgui" when compiling Xie language source code (Go language version).

如果谢语言主程序是加了-ldflags="-H windowsgui"的编译参数编译出来的，则通过其编译谢语言代码后的可执行程序，也将没有命令行窗口，结合GUI编程，完全可以制作出标准的图形界面程序。如何编译谢语言代码，可以参见后面文档中说明。

If the main program of Xielang is compiled with the compilation parameter of -ldflags="-H windowsgui", then the executable program compiled with Xielang code will also have no command line window. Combined with GUI programming, standard graphical interface programs can be produced. How to compile Xielang code can be explained in the following documents.

&nbsp;

##### - **制作一个登录框**（Create a login box）

本例继续介绍GUI编程，将实现一个常见的登录框，包含用户名和密码的输入框以及登录和关闭按钮，直接参看下面的代码（loginDialogGui.xie）：

This example continues to introduce GUI programming, implementing a common login box that includes input boxes for username and password, as well as login and close buttons. Please refer to the following code (loginDialogGui.xie):

```go
// 本例演示使用WebView2搭建一个登录对话框
//This example demonstrates using WebView2 to build a login dialog box

// 设定界面的HTML
// 其中Javascript代码中delegateDo函数是默认约定的使用setDelegate设置代理函数后与谢语言进行互通的函数
// 它接收一个字符串类型的输入参数，并输出一个字符串类型的输出参数
// 如果想传递多于一个的数据，可以用JSON进行数据的封装
// set HTML for GUI
// The delegateDo function in Javascript code is the default function to communicate with Xielang after using setDelegate instruction to set the delegate function in Xielang
// It receives an input parameter of string type and outputs an output parameter of string type
// If you want to transfer more than one data, you can use JSON for data encapsulation
assign $htmlT `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <title>Please login...</title>
</head>
<body >
	<div style="margin-top: 10px; margin-bottom: 10px;">
		<span>Please enter the user name and password to login...</span>
	</div>
	<div style="margin-top: 10px; margin-bottom: 10px;">
		<label for="userNameID" >User Name:&nbsp; </label><input id="userNameID" type=text />
	</div>
	<div style="margin-top: 10px; margin-bottom: 10px;">
		<label for="userNameID" >Password:&nbsp; </label><input id="passwordID" type=password />
	</div>
	<div>
		<button id="btnLoginID">Login</button>
		<button id="btnClose">Close</button>
	</div>

    <script>
        document.getElementById("btnLoginID").onclick = function() {
            var userNameT = document.getElementById("userNameID").value.trim();
			var passwordT = document.getElementById("passwordID").value.trim();

            // 调用代理函数与谢语言通信，传入参数并获取结果
            // 如果是使用setQuickDelegate指令设置的快速代理函数，则应该调用quickDelegateDo函数
            // Calling delegate functions to communicate with Xielang, passing in parameters and obtaining results
            // If the quick delegate function is set using the setQuickDelegate instruction, the quickDelegateDo function should be called
            let result = delegateDo(JSON.stringify({"userName": userNameT, "password": passwordT}));
            // let result = quickDelegateDo(JSON.stringify({"userName": userNameT, "password": passwordT}));

            result.then((r) => {
                // 弹框提示函数返回结果
                // show the result message
                alert("result: " + r);
            });

        };
 
        document.getElementById("btnClose").addEventListener("click", function() {
            // delegateCloseWindow函数是默认约定的关闭窗口的函数
            // The delegateCloseWindow function is the default convention for closing windows
            delegateCloseWindow();
        });

        document.addEventListener('DOMContentLoaded', function() {
            console.log("document loaded");
        });


    </script>
</body>
</html>
`

// 新建WebView2窗口，并指定宽、高，以及位置居中，并打开调试模式（可用右键检查）
// Create a new WebView2 window, specify the width, height, and center position, and open debug menu (right click to check)
mt $windowT $guiG newWindow "-title=Test" -width=600 -height=400 -center -debug

// 设置与界面之间的代理或快速代理对象
// 这里演示了4种调用代理函数的方法，推荐未注释的方法，但其他方法也可以选用
// Set delegate or quick delegate objects between the GUI and backend
// Here are four methods for calling delegate functions demonstrated. The uncommented method is recommended, but other methods can also be used

// 第一种方法：采用new指令创建快速代理函数，然后使用setQuickDelegate指令设置
// 快速代理对象使用inputL和outL变量来传递输入参数和输出参数
// 函数退出时使用exitL指令（这里带参数表示退出前将outL赋值为该参数代表的值）
// 快速代理函数需要在JavaScript代码中用quickDelegateDo函数来调用，这是个约定好的函数名字
// The first method is to use the new instruction to create a quick delegate function, and then use the setQuickDelegate instruction to set it
// Quick delegate objects use inputL and outL variables to pass input and output parameters
// When the function exits, use the exitL instruction (where a parameter indicates assigning outL to the value represented by the parameter before exiting)
// The quick delegate function needs to be called in JavaScript code using the quickDelegateDo function, which is a predetermined function name

// pln "method1"
// new $dele1 quickDelegate `
//     [] $resultL $inputL 0

//     pl "Result: %v" $resultL

//     // 快速代理函数返回前必须要有一个输出参数存入outL中
//     // 此处因为实际上无需返回参数，因此随便存入一个无用的数值
//     // There must be an output parameter stored in outL before the function returns
//     // Because there is actually no need to return parameters, a useless numerical value is randomly stored here
//     exitL $resultL
// `

// mt $rs $windowT setQuickDelegate $dele1

// 第二种方法：用new指令新建代理函数，然后用setDelegate指令来设置
// 与快速代理函数不同，代理函数将运行在不同的虚拟机中，相对更安全
// 代理函数与快速代理对象的区别是其中使用inputG和outG变量来传递输入参数和输出参数，并用exit指令退出
// 并且代理函数在JavsScript中用delegateDo来调用，而快速代理函数是用quickDelegateDo来调用
// The 2nd method: Use the new instruction to create a new delegate function, and then use the setDelegate instruction to set it
// Unlike quick delegate functions, delegate functions will run on different virtual machines and are relatively safer
// The difference between delegate functions and quick delegate objects is that they use inputG and outG variables to pass input and output parameters, and exit with the exit command
// And the delegate function is called with delegateDo in JavsScript, while the fast delegate function is called with quickDelegateDo

pln "method2"
new $dele1 delegate `
    [] $resultT $inputG 0

    pl "Result: %v" $resultT

    // 代理函数返回前必须要有一个输出参数存入outG中
    // 此处因为实际上无需返回参数，因此随便存入一个无用的数值
    // There must be an output parameter stored in outG before the function returns
    // Because there is actually no need to return parameters, a useless numerical value is randomly stored here
    exit $resultT
`

mt $rs $windowT setDelegate $dele1

// 第三种方法：直接使用字符串设置快速代理函数
// The third method: directly use a string to set up a quick delegate function

// pln "method3"
// mt $rs $windowT setQuickDelegate `

//     [] $resultT $inputL 0

//     pl "Result: %v" $resultT

//     exitL $resultT

// `

// // 第四种方法：直接使用字符串设置代理函数
// // The 4th method: directly use a string to set the delegate function

// pln "method4"
// mt $rs $windowT setDelegate `

//     [] $resultT $inputG 0

//     // ifEqual $resultT "closeWindow" :+1 :next1
//     //     mt $drop $windowT close
//     //     exit $resultT

//     pl "Result: %v" $resultT

//     exit $resultT

// `

// 设置窗口中使用的HTML
// Set the HTML used in the window
mt $rs $windowT setHtml $htmlT

plo $rs $windowT

// 运行图形界面
// 这是阻塞执行的，窗口被关闭才会执行后面的语句
// Run the GUI window
// This is blocking execution, and subsequent statements will only be executed when the window is closed
mt $rs $windowT show

plo $rs

// 关闭图形窗口
// Close the GUI window
mt $rs $windowT close

exit

```

运行效果如下图所示：

The operation effect is shown in the following figure:

![截图](http://xie.topget.org/example/xie/snap/snap10.png)

可以看出，点击登录按钮后，接口代理函数将输出一个JSON格式的包含输入的用户名和密码的字符串，可以用于后续处理。

It can be seen that after clicking the login button, the interface delegate function will output a JSON formatted string containing the input username and password, which can be used for subsequent processing.

&nbsp;

#### 编译运行谢语言代码（Compile and run Xielang code）

谢语言支持简单的编译运行，但仅相当于将主程序和代码打包成一个可执行文件，方便分发并起到简单加密代码的作用。例如要编译一个名为hello.xie的文件，用下面的命令：

Xielang supports simple compilation and operation, but it is only equivalent to packaging the main program and code into an executable file to facilitate distribution and play the role of simple encryption code. For example, to compile a file named hello.xie, use the following command:

```shell
xie -compile hello.xie -output=hello.exe
```

执行后将在当前目录下生成hello.exe的可执行文件（Linux类似），如果不指定output参数，则默认生成可执行文件名为output.exe。

After execution, the executable file of hello.exe will be generated in the current directory (similar to Linux). If the output parameter is not specified, the default executable file name is output.exe.

如果谢语言主程序是加了-ldflags="-H windowsgui"的编译参数编译出来的，则通过其编译后的可执行程序，也将没有命令行窗口，结合GUI编程，完全可以制作出标准的图形界面程序。

If the main program of Xielang is compiled with the compilation parameter of -ldflags="-H windowsgui", the compiled executable program will also have no command line window. In combination with GUI programming, standard graphical interface programs can be produced.

&nbsp;

#### 内置指令/命令/函数参考（Built-in instruction/command/function reference）

请参考[这里](instr.md)。或者目前暂时请参看代码中的InstrNameSet数据结构的代码注释，后面文档会慢慢补齐。

For the moment, please refer to the code comments of the InstrNameSet data structure in the code, and the following documents will be supplemented slowly.

&nbsp;

#### 内置对象参考（Built-in object reference）

目前暂时请参看代码中的各个Xie...对象（如XieString）的代码内文档说明。

For now, please refer to each Xie... object in the in-code document description of object (such as XieString).

&nbsp;

#### 杂项说明（Miscellaneous description）

&nbsp;

&nbsp;

##### - **指令的参数**（Parameter of instruction）

谢语言中的指令，可以没有任何参数（0个参数），即不需要输出也不需要输入参数，例如pass。也有可能只有一个结果参数，例如getNowStr，此时可以省略结果参数以表示将结果存入全局变量\$tmp。当然，也有可能既有结果参数，也有其他一个或多个输入参数。当输入参数是可变个数的时候，结果参数是不可省略的。输入参数固定的情况下，一般结果参数也可以省略来表示压栈。一般情况下，为了避免混淆，对于有结果参数的指令，建议总是写上结果参数。

Instructions in Xielang can have no parameters (0 parameters), that is, no output or input parameters are required, such as pass. It is also possible that there is only one result parameter, such as getNowStr. At this time, the result parameter can be omitted to indicate that the result will be saved to the global variable \$tmp. Of course, it is also possible to have both result parameters and one or more other input parameters. When the input parameters are variable, the result parameters cannot be omitted. When the input parameter is fixed, the general result parameter can also be omitted to indicate the stack pressing. In general, in order to avoid confusion, it is recommended to always write the result parameters for instructions with result parameters.

* 注：少数指令可以带有多个结果参数，例如getIter。

*Note: A few instructions can have multiple result parameters, such as getIter.

&nbsp;

##### - **行末注释**（Comment at the end of the line）

谢语言中的注释是不支持行内注释的，只能单独写在一行中。但在参数个数固定的指令中，如果显式写出了结果参数，此时可以利用该指令将忽略后面的其他参数的特性，来写上该行的注释。

Xielanguage do not support inline comments and comments can only be written in a single line. However, in an instruction with a fixed number of parameters, if the result parameter is explicitly written, you can use the feature that the instruction will ignore other parameters later to write a comment on the line.

&nbsp;

##### - **自动执行**（Auto-run scripts）

谢语言中主程序运行时，如果不指定要执行的脚本文件，同时当前目录下含有名字类似auto*.xie（例如auto.xie、auto01.xie等）的脚本文件时，将按文件名顺序依次执行这些脚本文件。这在分发程序时会比较有用，使用者可以直接鼠标双击谢语言主程序即可执行开发者编写的脚本，只要这些脚本与谢语言主程序在相同目录下并符合上述命名规则。

When the main program of Xielang(i.e. xie.exe in Windows, or xie in Linux) is running, if the script file to be executed is not specified, and the current directory contains script files with names similar to auto*.xie (such as auto.xie, auto01.xie, etc.), these script files will be executed in order of file names. This will be useful when distributing programs. Users can directly double-click the main program of Xielang to execute scripts written by developers, as long as these scripts are in the same directory as the main program of Xielang and conform to the above naming rules.

&nbsp;

##### - **从剪贴板执行代码**（Run Xielang code from clipboard）

谢语言主程序执行时，如果加上-clip参数，将从剪贴板读取代码然后执行。

When the main program of Xielang is executed, if the "-clip" parameter is added, the code will be read from the clipboard and then executed.

&nbsp;

##### - **指令参数中引号的位置**（The position of quotation marks in instruction parameters）

谢语言的指令中，可以使用双引号、单引号或反引号括起内容，包括字符串、表达式等，注意，一个参数中的引号就算从中间开始也会进行匹配，因此，下面几种引号内的情况都被认为是一个参数而非多个：

In Xielang instructions, double quotation marks, single quotation marks, or back quotation marks can be used to enclose content, including strings, expressions, etc. Note that the quotation marks in a parameter will match even from the middle. Therefore, the following situations within quotation marks are considered as one parameter rather than multiple:

```go
= $t4 #t`2022-08-06 11:22:00.019`

adds $s1 "This is a pie."

excelSetCell $drop $excelT "sheet1" @`{spr $tmp "B%d" $y}` {$objT,姓名}

excelSaveAs $result $excelT @`clresExport + {nowStrCompact} + .xlsx`

...
```

&nbsp;

##### - **fastCall指令调用的快速函数代码中使用+1等虚拟标号**（The "+1" virtual label used in the fast function code called by the fastCall instruction）

谢语言中fastCall指令调用的快速函数代码中，应避免使用+1、+3等虚拟标号，尽量使用:next1这种标准标号，但可以使用:+1，:+3这样形式的虚拟标号。

In the fast function code called by the fastCall instruction in Xielang, virtual labels such as +1 and +3 should be avoided, and the standard labels such as :next1 should be used as much as possible, but virtual labels such as :+1, :+3 can be used.

&nbsp;


&nbsp;

#### 性能方面的考虑（Performance considerations）

谢语言的目标是使用简单的语法结构减少脚本语言的语法解析开销以便提升速度，并且通过广泛使用内置指令避免使用速度很慢的反射。具体速度评估可以参考例子代码中的斐波那契数列产生的两个例子（递归方式fix.xie和循环方式fibFlat.xie）。

The goal of Xielang is to use simple syntax structure to reduce the syntax parsing cost of script language in order to improve the speed, and avoid the use of slow reflection through extensive use of built-in instructions. For specific speed evaluation, please refer to the two examples generated by Fibonacci sequence in the example code (the recursive method fix.xie and the circular method fibFlat.xie).

&nbsp;

#### 嵌入式使用谢语言（以虚拟机的方式在其他语言中调用）（Embedded Xielang in other languages）

- **在Go语言（Golang）中如何嵌入**：请参看cmd目录下的main.go，这是谢语言的主程序，里面既是以嵌入式的方法创建谢语言虚拟机并执行代码的。

- **How to embed Xielang in Go language (Golang)**: please refer to main.go under cmd directory, which is the main program of Xielang. It is used to create Xielang virtual machine and execute code in an embedded way.

&nbsp;

#### 扩展谢语言（Extended Xielang）

扩展谢语言一般来说有两个方法：

Generally, there are two ways to extend Xielang:

- **增加内置指令**：请fork本库，参考xie.go中的源代码，参看各个指令的写法编写自己的新指令，然后编译出可执行代码即可。

- **Add built-in instructions**: Please fork this library, refer to the source code in xie.go, and write your own new instructions according to the writing method of each instruction, and then compile the executable code.

- **增加内置对象**：请fork本库，参考xie.go中的源代码，各个Xie...对象（如XieString）的代码内文档说明，重点是实现XieObject接口，然后编译出可执行代码即可。

- **Add built-in objects**: Please fork this library, refer to the source code in xie.go, and the in-code documentation of each XieObject object (such as XieString). The key point is to implement the XieObject interface, and then compile the executable code.

&nbsp;

#### 编译谢语言（Compile Xielang）

- 目前谢语言还在积极开发中，为方便起见，go.mod中所有“replace github.com/topxeq/tk v1.0.1 => ../tk”一行，以便使用本地的tk库代替在线的。因此编译时需要将github.com/topxeq/tk中的库git clone到本地，或者将go.mod中这一行去掉即可。

- At present, Xielang is still actively developing. For convenience, the "replace github.com/topxeq/tk v1.0.1=>../tk" line in go.mod are used to use local tk libraries instead of the online one. Therefore, during compilation, it is necessary to "git clone" the library from github.com/topxeq/tk locally, or remove this line from go.mod.

- 在Linux下如果出现类似“package gl was not found in the pkg-config search path.”的错误：请执行 apt install libgl1-mesa-dev 命令安装依赖库。

- If an error similar to 'package gl was not found in the pkg config search path.' appears under Linux, please execute the 'apt install libgl1 mesa dev' command to install the dependent library.

&nbsp;

#### 代码示例（Code examples）

*注：更多示例请参考cmd/scripts目录*

*Note: For more examples, please refer to the cmd/scripts directory of source repository*

- [三元操作符 ?（ternary operator '?'）](http://xie.topget.org/xc/c/xielang/example/operator3.xie)
- [MD5编码](http://xie.topget.org/xc/c/xielang/example/md5.xie)
- [命令行获取用户输入及密码输入](http://xie.topget.org/xc/c/xielang/example/input.xie)
- [文本加解密](http://xie.topget.org/xc/c/xielang/example/encryptText.xie)
- [二进制数据或文件加解密](http://xie.topget.org/xc/c/xielang/example/encryptFile.xie)
- [复制目录结构](http://xie.topget.org/xc/c/xielang/example/genFakeDirs.xie)
- [查找重复文件](http://xie.topget.org/xc/c/xielang/example/findDuplicateFiles.xie)
- [字节列表/数组的操作与16进制编解码](http://xie.topget.org/xc/c/xielang/example/hex.xie)
- [按二进制位计算](http://xie.topget.org/xc/c/xielang/example/bitwise.xie)
- [随机落点法计算圆周率π（Pi）](http://xie.topget.org/xc/c/xielang/example/calPi.xie)
- [数据库操作（SQLite3为例）](http://xie.topget.org/xc/c/xielang/example/sqlite.xie)
- [获取屏幕分辨率](http://xie.topget.org/xc/c/xielang/example/getScreenInfo.xie)
- [屏幕截图](http://xie.topget.org/xc/c/xielang/example/captureScreen.xie)
- [启动WEB服务分享文本文件供编辑](http://xie.topget.org/xc/c/xielang/example/editFileServer.xie)
- [通过SSH直接编辑远程服务器上的文件](http://xie.topget.org/xc/c/xielang/example/sshEdit.xie)

&nbsp;

#### 参与贡献者（Contributors）

1.  TopXeQ
2.  Topget
3.  陆满庭



