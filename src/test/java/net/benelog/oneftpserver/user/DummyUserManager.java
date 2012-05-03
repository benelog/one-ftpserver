package net.benelog.oneftpserver.user;

import java.util.Arrays;

import org.apache.ftpserver.ftplet.Authentication;
import org.apache.ftpserver.ftplet.AuthenticationFailedException;
import org.apache.ftpserver.ftplet.Authority;
import org.apache.ftpserver.ftplet.FtpException;
import org.apache.ftpserver.ftplet.User;
import org.apache.ftpserver.ftplet.UserManager;
import org.apache.ftpserver.usermanager.UsernamePasswordAuthentication;
import org.apache.ftpserver.usermanager.impl.BaseUser;
import org.apache.ftpserver.usermanager.impl.ConcurrentLoginPermission;
import org.apache.ftpserver.usermanager.impl.WritePermission;

/**
 * @author benelog
 */
public class DummyUserManager implements UserManager {

	private String homeDir;
	public DummyUserManager(String homeDir) {
		this.homeDir = homeDir;
	}

	@Override
	public User getUserByName(String username) throws FtpException {
		BaseUser user = new BaseUser();
		user.setName(username);
		user.setEnabled(true);
		user.setAuthorities(Arrays.<Authority>asList(
				new WritePermission(),
				new ConcurrentLoginPermission(0, 0)
			));
		return user;
	}

	@Override
	public String[] getAllUserNames() throws FtpException {
		return new String[0];
	}

	@Override
	public void delete(String username) throws FtpException {
	}

	@Override
	public void save(User user) throws FtpException {

	}

	@Override
	public boolean doesExist(String username) throws FtpException {
		return true;
	}

	@Override
	public User authenticate(Authentication authentication)
			throws AuthenticationFailedException {
		if (!(authentication instanceof UsernamePasswordAuthentication)) {
			return null;
		}

		UsernamePasswordAuthentication userInfo = (UsernamePasswordAuthentication) authentication;
		BaseUser user = new BaseUser();
		user.setName(userInfo.getUsername());
		user.setMaxIdleTime(500);
		user.setEnabled(true);
		user.setHomeDirectory(homeDir);
		user.setAuthorities(Arrays.<Authority>asList(
				new WritePermission(),
				new ConcurrentLoginPermission(0, 0)
			));
		return user;
	}

	@Override
	public String getAdminName() throws FtpException {
		return "admin";
	}

	@Override
	public boolean isAdmin(String username) throws FtpException {
		return false;
	}
}
