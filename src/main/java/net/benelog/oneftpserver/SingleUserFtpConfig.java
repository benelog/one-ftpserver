package net.benelog.oneftpserver;

import java.util.Properties;

import org.apache.commons.collections.MapUtils;

/**
 * @author benelog
 */
public class SingleUserFtpConfig {
	public static final int DEFAULT_PORT = 2121;
	public static final String ANONYMOUS_ID = "anonymous";
	
	private int port;
	private String id;
	private String password;
	private String home;
	private boolean ssl;
	private String passivePorts;
	
	public SingleUserFtpConfig(Properties configParams){
		this.port = MapUtils.getIntValue(configParams,"port", DEFAULT_PORT);
		this.ssl = MapUtils.getBooleanValue(configParams,"ssl", false);
		this.id = configParams.getProperty("id", ANONYMOUS_ID);
		this.password = configParams.getProperty("password", id);
		this.home = configParams.getProperty("home", System.getProperty("user.dir"));
		this.passivePorts = configParams.getProperty("passivePorts");
	}
	
	public boolean isDefaultPort() {
		return this.port == DEFAULT_PORT;
	}
	
	public boolean isAnonymousId(){
		return ANONYMOUS_ID.equals(this.id);
	}
	
	public int getPort() {
		return port;
	}
	public String getId() {
		return id;
	}
	public String getPassword() {
		return password;
	}

	public String getHome() {
		return home;
	}

	public boolean isSsl() {
		return ssl;
	}

	public String getPassivePorts() {
		return passivePorts;
	}
	
}
