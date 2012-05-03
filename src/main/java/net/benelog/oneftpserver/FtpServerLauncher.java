package net.benelog.oneftpserver;

import java.io.File;

import org.apache.ftpserver.DataConnectionConfiguration;
import org.apache.ftpserver.DataConnectionConfigurationFactory;
import org.apache.ftpserver.FtpServer;
import org.apache.ftpserver.FtpServerFactory;
import org.apache.ftpserver.ftplet.FtpException;
import org.apache.ftpserver.ftplet.UserManager;
import org.apache.ftpserver.listener.Listener;
import org.apache.ftpserver.listener.ListenerFactory;
import org.apache.ftpserver.ssl.SslConfiguration;
import org.apache.ftpserver.ssl.SslConfigurationFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import net.benelog.oneftpserver.util.NetworkUtils;

/**
 * @author benelog
 */
public class FtpServerLauncher {
	private static final Logger log = LoggerFactory.getLogger(FtpServerLauncher.class);
	private int port;
	private UserManager userManager;
	private String passivePorts;
	private File keyStoreFile;
	private String password;
	private FtpServer server;
	private boolean running = false;
	private Listener listener;
	private boolean ssl = false;

	public FtpServerLauncher(int port, String passivePorts, UserManager userManager, File keyStoreFile, String password) {
		this.port = port;
		this.passivePorts = passivePorts;
		this.userManager = userManager;
		this.keyStoreFile = keyStoreFile;
		this.password = password;
		this.ssl = true;
	}

	public FtpServerLauncher(int port, String passivePorts, UserManager userManager) {
		this.port = port;
		this.passivePorts = passivePorts;
		this.userManager = userManager;
	}

	public void start() throws FtpException {
		server = createServer();
		synchronized(this){
			if(!this.running){
				server.start();
				this.running = true;
			}
		}
		log.info("Starting FTP sever... ({}:{})", NetworkUtils.getLocalhostIp(), port);
		addShutdownHook(server);
	}

	private FtpServer createServer() {
		this.listener = createListener();

		FtpServerFactory serverFactory = new FtpServerFactory();
		serverFactory.addListener("default", listener);
		serverFactory.setUserManager(this.userManager);
		return serverFactory.createServer();
	}
	
	private Listener createListener() {
		ListenerFactory listenerFactory = new ListenerFactory();
		listenerFactory.setPort(this.port);

		// config ssl
		if(this.ssl){
			SslConfiguration sslConfig = createSslConfig();
			listenerFactory.setSslConfiguration(sslConfig);
			listenerFactory.setImplicitSsl(true);
		}
		
		// config passive ports
		DataConnectionConfiguration conConfig = createConnectionConfig();
		listenerFactory.setDataConnectionConfiguration(conConfig);

		return listenerFactory.createListener();
	}

	
	private SslConfiguration createSslConfig() {
		SslConfigurationFactory ssl = new SslConfigurationFactory();
		ssl.setKeystoreFile(this.keyStoreFile);
		ssl.setKeystorePassword(this.password);
		return ssl.createSslConfiguration();
	}

	private DataConnectionConfiguration createConnectionConfig() {
		DataConnectionConfigurationFactory conConfigFactory = new DataConnectionConfigurationFactory();
		
		if (this.passivePorts!=null) {
			conConfigFactory.setPassivePorts(this.passivePorts);
		}
		return conConfigFactory.createDataConnectionConfiguration();
	}


	private void addShutdownHook(final FtpServer engine) {
		Runnable shutdownHook = new Runnable() {
			public void run() {
				log.info("Stopping server by shutdown hook");
				engine.stop();
			}
		};

		Runtime runtime = Runtime.getRuntime();
		runtime.addShutdownHook(new Thread(shutdownHook));
	}
	
	public synchronized void stop(){
		server.stop();
	}

	public FtpServer getFtpServer() {
		return this.server;
	}
}
