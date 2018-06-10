# SimpleMIPSComplier

将MIPS汇编语言转换成二进制机器码，并且生成带有十六进制的表格文档

（为了防止表中的0被Excel自动省略，我在每个格后面都加了一个&符号，如果需要去掉使用替换就可以）



## 使用方法

```shell
mipsc.exe source.asm res.txt
```

`source.asm`为汇编源文件

`res.txt`为需要生成的目标文件

`doc.csv`为附加生成的表格文件



`bin`里面由样例参考