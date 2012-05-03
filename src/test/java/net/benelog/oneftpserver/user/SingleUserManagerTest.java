package net.benelog.oneftpserver.user;

import static org.hamcrest.CoreMatchers.*;
import static org.junit.Assert.*;

import org.apache.ftpserver.ftplet.AuthenticationFailedException;
import org.apache.ftpserver.ftplet.User;
import org.apache.ftpserver.usermanager.AnonymousAuthentication;
import org.apache.ftpserver.usermanager.UsernamePasswordAuthentication;
import org.junit.Test;


/**
 * @author benelog
 */
public class SingleUserManagerTest {
	String home = "/";
	
	@Test
	public void testAnonymous() throws AuthenticationFailedException{
		String id = "anonymous";
		SingleUserManager userManager = new SingleUserManager(id, home);
		
		AnonymousAuthentication authentication = new AnonymousAuthentication();
		User user = userManager.authenticate(authentication);
		
		assertThat(user, is(notNullValue()));
		assertThat(user.getName(), is(id));
	}
	
	@Test(expected=AuthenticationFailedException.class)
	public void testAnonymousFail() throws AuthenticationFailedException{
		String id = "benelog";
		String password = "pwpw1";
		SingleUserManager userManager = new SingleUserManager(id,password, home);
		
		AnonymousAuthentication authentication = new AnonymousAuthentication();
		User user = userManager.authenticate(authentication);
		
		assertThat(user, is(notNullValue()));
		assertThat(user.getName(), is(id));
	}
	
	@Test
	public void testNormalUser() throws AuthenticationFailedException{
		String id = "benelog";
		String password = "pwpw1";
		SingleUserManager userManager = new SingleUserManager(id,password, home);
		
		UsernamePasswordAuthentication authentication = new UsernamePasswordAuthentication(id, password);
		User user = userManager.authenticate(authentication);
		
		assertThat(user, is(notNullValue()));
		assertThat(user.getName(), is(id));
	}

	@Test(expected=AuthenticationFailedException.class)
	public void testInvalidId() throws AuthenticationFailedException{
		String id = "benelog";
		String password = "pwpw1";
		SingleUserManager userManager = new SingleUserManager(id, password, home);
		
		UsernamePasswordAuthentication authentication = new UsernamePasswordAuthentication("bene", password);
		userManager.authenticate(authentication);
		
	}

	@Test(expected=AuthenticationFailedException.class)
	public void testInvalidPassword() throws AuthenticationFailedException{
		String id = "benelog";
		String password = "pwpw1";
		SingleUserManager userManager = new SingleUserManager(id, password, home);
		
		UsernamePasswordAuthentication authentication = new UsernamePasswordAuthentication(id, "1111");
		userManager.authenticate(authentication);
	}
}
