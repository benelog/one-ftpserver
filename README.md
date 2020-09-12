Using One-FTPServer
=========
# Features

[![Join the chat at https://gitter.im/benelog/one-ftpserver](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/benelog/one-ftpserver?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
- Provides a simple way to run [Apache FTP sever](http://mina.apache.org/ftpserver/).
	- No installation. Packaged in one executable jar file.
	- No configuration file. All configuration is supported by command line arguments.
- Supports anonymous login by default.
- Supports FTP and FTPS(FTP over SSL(Secure Sockets Layer))

# Download
You can download an executable jar file from <https://github.com/benelog/one-ftpserver/releases/download/0.1/one-ftpserver.jar>

	wget https://github.com/benelog/one-ftpserver/releases/download/0.1/one-ftpserver.jar
    
# Execute
If you execute a jar file with no argument, it will be running with default configuration.

	java -jar one-ftpserver.jar

You can specify parameters for configurations by command line arguments with 'key=value' format.

	java -jar one-ftpserver.jar port=10021 id=benelog password=1234
	
   
Currently, 6 parameters are supported as following.

	java -jar one-ftpserver.jar port=10021 passivePorts=10125-10199 ssl=true id=benelog password=1234 home=/home/benelog/programs

After the execution, configuration parameters are printed to the standard output.

	FTP sever started : ftp://127.0.1.1:2003 
	# Configuration parameters
	- ssl : false 
	- port : 2003
	- passivePorts : 
	- id : benelog
	- password : 1234
	- home : /home/benelog/workspace/one-ftpserver

# Confiugration Parameters
- ssl : true/false. The deafult value is 'false'. If it is 'true', FTPS will be used. 
- port : a control port for FTP/FTPS. The default value is 2121.
- passivePorts : ports for passive data connection. You can define a range of ports like '10125-10199'  More examples are on <http://mina.apache.org/ftpserver/configure-passive-ports.html>
- id : username for login. The default value is 'anonymous'. If you don't specify this parameter, anonymous login will be activated.
- password : password for login.
- home : home directory of user. The default value is working directory of the FTP server.
