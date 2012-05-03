package net.benelog.oneftpserver;

import static org.junit.Assert.*;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

import org.apache.commons.net.PrintCommandListener;
import org.apache.commons.net.ftp.FTPClient;
import org.apache.commons.net.ftp.FTPSClient;
import org.apache.ftpserver.ftplet.FtpException;
import org.junit.After;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.TemporaryFolder;


/**
 * @author benelog
 */
public class CommanderFtpSslTest {
	@Rule
	public TemporaryFolder folder = new TemporaryFolder();
	final ExecutorService threadPool = Executors.newFixedThreadPool(1);
	
	String home;
	FTPClient client = new FTPSClient(true);

	@Before
	public void setUp() {
		client.addProtocolCommandListener(new PrintCommandListener(System.out));
		File uploadBase = folder.getRoot();
		home = uploadBase.getAbsolutePath();
	}
	

	@After
	public void tearDown() throws IOException, InterruptedException {
		client.disconnect();
		Commander.requestStop();
	}

	@Test
	public void login() throws Exception {
		startServer(new String[]{"port=3131","ssl=true", "id=benelog","password=1234", "home=" + home});
		client.connect("127.0.0.1", 3131);

		// login
		boolean authorized = client.login("benelog", "1234");
		assertTrue(authorized);
	}
	
	@Test
	public void loginFail() throws Exception {
		startServer(new String[]{"port=3131","ssl=true","id=benelog","password=1234", "home=" + home});
		client.connect("127.0.0.1", 3131);

		// login
		boolean authorized = client.login("benelog", "13234");
		assertFalse(authorized);
	}
	
	@Test
	public void loginAnonymousAndUpload() throws Exception {
		startServer(new String[]{"ssl=true", "home=" + home});
		client.connect("127.0.0.1", 2121);

		// login
		boolean authorized = client.login("anonymous", "benelog@gmail.com");
		assertTrue(authorized);
		
		// upload
		String fileName = "test.txt";
		InputStream content = openStreamFromClassPath(fileName);
		boolean stored = client.storeFile(fileName, content);
		assertTrue(stored);
		String savedFile = folder.getRoot().getAbsolutePath() + "/" + fileName;
		assertTrue(new File(savedFile).exists());
	}

	private InputStream openStreamFromClassPath(String fileName) throws IOException {
		return this.getClass().getClassLoader().getResource(fileName).openStream();
	}
	
	private void startServer(final String[] args) throws FtpException {
		threadPool.execute(new Runnable(){
			@Override
			public void run() {
				try {
					Commander.main(args);
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
}
