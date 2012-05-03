package net.benelog.oneftpserver;

import java.io.File;
import java.io.IOException;
import java.net.URL;

import org.apache.commons.io.FileUtils;
import org.apache.ftpserver.ftplet.UserManager;

import net.benelog.oneftpserver.user.SingleUserManager;

/**
 * @author benelog
 */
public class FtpServerLauncherFactory {
	private String keyStoreFileName;
	private String keyPassword;

	public FtpServerLauncherFactory(String keyStoreFileName, String keyPassword){
		this.keyStoreFileName = keyStoreFileName;
		this.keyPassword = keyPassword;
	}

	public FtpServerLauncher createLauncher(SingleUserFtpConfig config) {
		UserManager userManager = createUserManager(config);
		
		if(config.isSsl()) {
			File keyStoreFile = getKeyStoreFile(this.keyStoreFileName);
			return new FtpServerLauncher(config.getPort(), config.getPassivePorts(), userManager, keyStoreFile, this.keyPassword);
		}
		return new FtpServerLauncher(config.getPort(), config.getPassivePorts(), userManager);
	}

	protected UserManager createUserManager(SingleUserFtpConfig config) {
		if(config.isAnonymousId()) {
			return new SingleUserManager(config.getId(), config.getHome());
		} 
		return new SingleUserManager(config.getId(), config.getPassword(), config.getHome());
	}

	protected File getKeyStoreFile(String fileName) {
		URL source = FtpServerLauncherFactory.class.getClassLoader().getResource(fileName);
		File destination = new File(fileName);
		try {
			FileUtils.copyURLToFile(source, destination);
		} catch (IOException e) {
			throw new IllegalStateException("fail to copy key store file :" + fileName, e);
		}
		return destination;
	}
}
