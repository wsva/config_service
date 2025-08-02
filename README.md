ConfigService provides configurations to every machine.

# Architecture
Root node:
+ generate complete config data from multiple sources
+ provide config query service
+ send to branch nodes in zip-file format
+ all files are encrypted

Branch Node:
+ read data from zip-file from root node
+ provide config query service
+ send to other branch nodes in zip-file format

# Structure of the Directory Source
When generating the config data, only the contents in index.json are used.

Each type of config is stored in its own separate subdirectory.

# The order of getting config
+ 1, match ip exactly
+ 2, match ip with regexp
+ 3, public

# design notes
1，为什么需要使用MongoDB？直接使用文件不是更简单吗？
+ 1）有的配置，直接使用文件更简单，那就用文件，不必再存到MongoDB中。
+ 2）有的配置，存在MongoDB中更便于统计。毕竟，在分散的文件中，想要系统地、整体地查看、统计，就不方便了。
+ 3）有些配置文件我们既希望保持简洁，又想有充分地注释说明，用MongoDB就更好，可以在MongoDB中写注释，但是生成的缓存保持简洁。
+ 4）ConfigService支持多种方式的缓存，想用那种就用那种，哪种方便就用哪种。

2，ConfigService的体系
+ 1）有一个ConfigService扮演root角色，负责从多个渠道生成最完整的config data。
+ 2）其他的ConfigService扮演branch角色，只是获取root分发的zip文件。
+ 3）root向所有的branch分发zip打包的config data。
+ 4）data中的文件全部加密，不管是否涉及密码。

3，目录源的结构
+ 1）目录下有一个索引文件，数据类型为ConfigIndex
+ 2）每种类型的config单独存放在一个目录，并添加到索引文件
+ 3）生成config时，只会根据索引文件生成config

4，MongoDB源的结构
+ 1）MongoDB中有一个索引Collection
+ 2）根据索引Collection找到数据所在Collection
+ 3）数据Collection可以是视图，是对真实数据Collection的抽象
+ 4）因为有这样一层抽象，可以略掉部分真实数据中的字段，因此方便添加注释
