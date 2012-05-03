package net.benelog.oneftpserver;

import static org.junit.Assert.*;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.net.URL;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

import org.apache.commons.net.PrintCommandListener;
import org.apache.commons.net.ftp.FTPSClient;
import org.apache.ftpserver.ftplet.FtpException;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.TemporaryFolder;

import net.benelog.oneftpserver.user.DummyUserManager;

/**
 * @author benelog
 */
public class FtpServerLauncherTest {
	@Rule
	public TemporaryFolder folder = new TemporaryFolder();
	
	int port = 21101;
	FtpServerLauncher launcher;
	FTPSClient client = new FTPSClient(true);

	@Before
	public void setUp() {
		client.addProtocolCommandListener(new PrintCommandListener(System.out));
	}

	@Test
	public void loginAndUpload() throws Exception {
		startServer();
		client.connect("127.0.0.1", this.port);

		// login
		boolean authorized = client.login("test1", "");
		assertTrue(authorized);
		
		// upload
		String fileName = "test.txt";
		InputStream content = openStreamFromClassPath(fileName);
		boolean stored = client.storeFile(fileName, content);
		assertTrue(stored);

		String savedFile = folder.getRoot().getAbsolutePath() + "/" + fileName;
		assertTrue(new File(savedFile).exists());
		
		client.disconnect();
	}

	private InputStream openStreamFromClassPath(String fileName) throws IOException {
		return this.getClass().getClassLoader().getResource(fileName).openStream();
	}
	
	
	public void startServer() throws FtpException {
		configServer();
		ExecutorService threadPool = Executors.newFixedThreadPool(2);
		threadPool.execute(new Runnable(){
			@Override
			public void run() {
				try {
					launcher.start();
				} catch (FtpException e) {
					e.printStackTrace();
				}
			}
		});
		try {
			TimeUnit.SECONDS.sleep(1);
		} catch (InterruptedException e) {
			//ignore
		}
	}

	private void configServer() throws FtpException {
		String passivePorts = "10125-10199";
		String password = "pass0101";
		String keyStoreFileName = "ftpkeystore.jks";
		File uploadBase = folder.getRoot();
		String accessHome = uploadBase.getAbsolutePath();

		File keyStoreFile = readFromClassPath(keyStoreFileName);

		launcher = new FtpServerLauncher(this.port, passivePorts, new DummyUserManager(accessHome),keyStoreFile, password);
	}

	private File readFromClassPath(String fileName) {
		URL resource = this.getClass().getClassLoader().getResource(fileName);
		return new File(resource.getFile());
	}

}
