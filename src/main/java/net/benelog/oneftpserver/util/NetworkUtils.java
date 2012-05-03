/*
 * @(#)NetworkUtils.java $version 2012. 5. 3.
 *
 * Copyright 2007 NHN Corp. All rights Reserved. 
 * NHN PROPRIETARY/CONFIDENTIAL. Use is subject to license terms.
 */

package net.benelog.oneftpserver.util;

import java.net.InetAddress;
import java.net.UnknownHostException;

/**
 * @author benelog
 */
public class NetworkUtils {
	public static String getLocalhostIp() {
		InetAddress localhost;
		try {
			localhost = InetAddress.getLocalHost();
		} catch (UnknownHostException e) {
			throw new IllegalStateException("fail to get localhost ip", e);
		}
		return localhost.getHostAddress();
	}
}
