def map = [
	//http相关设置
	"http.port":7070,
	"http.minThread":8,
	"http.maxThread":200,
	"http.threadIdleTimeout":60000,
	"http.acceptQueueSize":1200,
	"http.idleTimeout":60000,
	"http.applicationName":"das",
	
	//cookie相关设置
	"cookie.usingCookie":true,
	"cookie.httpOnly":true,
	"cookie.secureCookies":true,
	
	//jdbc相关设置
	"jdbc.connection.driverClassName":"com.mysql.jdbc.Driver",
	"jdbc.connection.url":"jdbc:mysql://mysql:3306/das?autoReconnect=true&failOverReadOnly=false",
	"jdbc.connection.username":"root",
	"jdbc.connection.password":"root",
	
	//connection pool set
	"proxool.testBeforeUse":"false",
	"proxool.simultaneousBuildThrottle":"20",
	"proxool.minimumConnectionCount":"10",
	"proxool.maximumConnectionCount":"50"
];

return map
