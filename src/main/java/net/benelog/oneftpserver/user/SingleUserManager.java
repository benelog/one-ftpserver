package net.benelog.oneftpserver.user;

import java.util.Arrays;

import org.apache.ftpserver.ftplet.Authentication;
import org.apache.ftpserver.ftplet.AuthenticationFailedException;
import org.apache.ftpserver.ftplet.Authority;
import org.apache.ftpserver.ftplet.FtpException;
import org.apache.ftpserver.ftplet.User;
import org.apache.ftpserver.ftplet.UserManager;
import org.apache.ftpserver.usermanager.AnonymousAuthentication;
import org.apache.ftpserver.usermanager.UsernamePasswordAuthentication;
import org.apache.ftpserver.usermanager.impl.BaseUser;
import org.apache.ftpserver.usermanager.impl.ConcurrentLoginPermission;
import org.apache.ftpserver.usermanager.impl.TransferRatePermission;
import org.apache.ftpserver.usermanager.impl.WritePermission;

/**
 * @author benelog
 */
public class SingleUserManager implements UserManager {
	private String homeDir;
	private String id;
	private String password;
	private boolean anonymous;

	public SingleUserManager(String id, String password, String homeDir) {
		this.id = id;
		this.password = password;
		this.homeDir = homeDir;
	}
	
	public SingleUserManager(String id, String homeDir) {
		this.id = id;
		this.homeDir = homeDir;
		this.anonymous = true;
	}

	@Override
	public User getUserByName(String id) throws FtpException {
		if (this.id.equals(id)){
			return createBaseUser(id);
		}
        	return null;
	}

	@Override
	public String[] getAllUserNames() throws FtpException {
		return new String[]{this.id};
	}

	@Override
	public void delete(String username) throws FtpException {
	}

	@Override
	public void save(User user) throws FtpException {
	}

	@Override
	public boolean doesExist(String id) throws FtpException {
		return id.equals(this.id);
	}

	@Override
	public User authenticate(Authentication authentication)
			throws AuthenticationFailedException {

		if ((authentication instanceof AnonymousAuthentication)) {
			if(this.anonymous) {
				return createBaseUser(this.id);
			}
           		throw new AuthenticationFailedException("anonymous user not allowed");
		}

		UsernamePasswordAuthentication userInfo = (UsernamePasswordAuthentication) authentication;
		String loginId = userInfo.getUsername();

		String loginPassword = userInfo.getPassword();
		
		if (!(this.id.equals(loginId))) {
            		throw new AuthenticationFailedException("invalid username");
		}
		
		if(this.password.equals(loginPassword)) {
			return createBaseUser(loginId);
		}
		throw new AuthenticationFailedException("");
	}

	private User createBaseUser(String loginId) {
		BaseUser user = new BaseUser();
		user.setName(loginId);
		user.setMaxIdleTime(500);
		user.setEnabled(true);
		user.setHomeDirectory(this.homeDir);
		user.setAuthorities(Arrays.<Authority>asList(
				new WritePermission(),
				new TransferRatePermission(0, 0),
				new ConcurrentLoginPermission(0, 0)
			));
		return user;
	}

	@Override
	public String getAdminName() throws FtpException {
		return this.id;
	}

	@Override
	public boolean isAdmin(String id) throws FtpException {
		return this.id.equals(id);
	}
}
