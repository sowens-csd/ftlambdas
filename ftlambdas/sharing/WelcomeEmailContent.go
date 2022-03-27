package sharing

import "github.com/sowens-csd/ftlambdas/awsproxy"

// EnglishNewUserWelcome the english content for the new user welcome email
func EnglishNewUserWelcome() awsproxy.EmailContent {
	return awsproxy.EmailContent{
		Subject: "Welcome to Folktells!",
		HTMLBody: `
		<!DOCTYPE html
		PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
	  <html xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml"
		xmlns:o="urn:schemas-microsoft-com:office:office">
	  
	  <head>
		<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<title>Welcome to Folktells!</title>
		<style type="text/css">
		  @import url('https://fonts.googleapis.com/css?family=Open+Sans&display=swap');
	  
		  table {
			font-family: 'Open Sans', Roboto, Helvetica, Verdana, Arial, sans-serif;
			font-size: 16px;
		  }
	  
		  @media only screen and (max-width: 650px) {
			body {
			  padding: 10px !important;
			}
	  
			.inner {
			  padding-left: 15px !important;
			  padding-right: 15px !important;
			}
	  
			.container {
			  width: 100% !important;
			}
	  
			.half-block {
			  display: block !important;
			}
	  
			.half-block tr {
			  display: block !important;
			}
	  
			.half-block td {
			  display: block !important;
			}
	  
			.half-block__image {
			  width: 100% !important;
			}
	  
			.half-block__content {
			  width: 100% !important;
			  box-sizing: border-box;
			  padding: 25px 15px 25px 15px !important;
			}
		  }
		</style>
		<!--[if gte mso 9]><xml>
				  <o:OfficeDocumentSettings>
					  <o:AllowPNG/>
					  <o:PixelsPerInch>96</o:PixelsPerInch>
				  </o:OfficeDocumentSettings>
			  </xml><![endif]-->
	  </head>
	  
	  <body style="padding: 0; margin: 0;" bgcolor="#eeeeee">
		<span style="color:transparent !important; 
			  overflow:hidden !important; 
			  display:none !important; 
			  line-height:0px !important; 
			  height:0 !important; 
			  opacity:0 !important; 
			  visibility:hidden !important; 
			  width:0 !important; 
		  mso-hide:all;">&#x1F389; Share your stories with family & friends</span>
		<table class="full-width-container" border="0" cellpadding="0" cellspacing="0" width="100%" bgcolor="#eeeeee"
		  style="width: 100%; height: 100%; padding: 30px 0 30px 0;">
		  <tr>
			<td align="center" valign="top">
			  <table class="container" border="0" cellpadding="0" cellspacing="0" width="650" bgcolor="#ffffff"
				style="width: 650px;">
				<tr>
				  <td align="center" valign="top">
					<table class="container header" border="0" cellpadding="0" cellspacing="0" width="100"
					  style="width: 100%;" bgcolor="#364778">
					  <tr>
						<td style="padding: 20px 20px 20px 20px; border-bottom: solid 1px #eeeeee;" align="left">
						  <a target="_blank" href="https://folktells.com/" style="text-decoration: none;">
							<img src="https://folktells.com/wp-content/uploads/2020/01/light-text@2x.png"
							  style="width:100%; max-width:151px;" alt="Folktells" width="151" height="55" /></a>
						</td>
					  </tr>
					</table>
					<table class="container title-block" border="0" cellpadding="0" cellspacing="0" width="620"
					  style="width: 620px;">
					  <tr>
						<td class="hero-subheader__title inner" style="font-size: 30px; padding: 30px 0 15px 0;" align="left">
						  Welcome to Folktells!</td>
					  </tr>
	  
					  <tr>
						<td class="hero-subheader__content inner" style="font-size: 16px; line-height: 27px;" align="left">
						  We are excited to have you join the Folktells community.</td>
					  </tr>
					</table>
					<table class="container title-block" border="0" cellpadding="0" cellspacing="0" width="100%">
					  <tr>
						<td align="center" valign="top">
						  <table class="container" border="0" cellpadding="0" cellspacing="0" width="620"
							style="width: 620px;">
							<tr>
							  <td class="inner"
								style="border-bottom: solid 1px #eeeeee; padding: 35px 0 18px 0; font-size: 26px;"
								align="left">Getting Started</td>
							</tr>
						  </table>
						</td>
					  </tr>
					</table>
					<table class="container title-block" border="0" cellpadding="0" cellspacing="0" width="620"
					  style="width: 620px;padding-top:20px;">
					  <tr>
						<td class="hero-subheader__content inner" style="font-size: 16px; line-height: 27px;" align="left">
						  Visit our website to learn how to <a
							href="https://folktells.com/how-to/create-and-manage-stories/" target="_blank">create & share
							stories</a>, connect to shared albums from <a
							href="https://folktells.com/how-to/albums-from-google-photos/" target="_blank">Google Photos and
							iCloud</a>, <a href="https://folktells.com/how-to/select-calendars-to-display/"
							target="_blank">view events from your calendars</a> and more.</td>
					  </tr>
					</table>
					<table class="container title-block" border="0" cellpadding="0" cellspacing="0" width="620"
					  style="width: 620px;padding-top:20px;">
					  <tr>
						<td class="hero-subheader__content inner" style="font-size: 16px; line-height: 27px;" align="left">
						  Invite your friends and family to join Folktells and share stories & photos about the special
						  events, people & pets, and adventures that life brings your way.</td>
					  </tr>
					</table>
					<table class="container half-block" border="0" cellpadding="0" cellspacing="0" width="100%"
					  style="padding-top: 25px;padding-left:15px">
					  <tr>
						<td>
						  <table class="container" border="0" cellpadding="0" cellspacing="0" width="100%">
							<tr>
							  <td class="half-block__image" style="width: 300px;" width="300">
								<img alt="Create your group" style=" text-align: center;width:100%; max-width:300px;"
								  width="300" src="https://folktells.com/wp-content/uploads/2020/05/create-group@2x.jpg" />
							  </td>
							  <td class="half-block__content inner inner"
								style="width: 50%; padding: 0 25px 0 25px; font-size: 16px; line-height: 27px;">
								Create a Folktells group:<br />
								* from the menu, choose Sharing<br />
								* on the Groups tab, be sure you are signed in and that Sharing is ON<br />
								* tap New and enter the name of your group<br />
								* tap create<br />
								* expand your new group section<br />
								* invite your family & friends to join.
							  </td>
							</tr>
						  </table>
						</td>
					  </tr>
					</table>
	  
					<table class="container title-block" border="0" cellpadding="0" cellspacing="0" width="100%">
					  <tr>
						<td align="center" valign="top">
						  <table class="container" border="0" cellpadding="0" cellspacing="0" width="620"
							style="width: 620px;">
							<tr>
							  <td class="inner" style="padding: 45px 0 10px 0; font-size: 20px;" align="left">Find out more
							  </td>
							</tr>
						  </table>
						</td>
					  </tr>
					</table>
					<table class="container paragraph-block" border="0" cellpadding="0" cellspacing="0" width="100%">
					  <tr>
						<td align="center" valign="top">
						  <table class="container" border="0" cellpadding="0" cellspacing="0" width="620"
							style="width: 620px;">
							<tr>
							  <td class="paragraph-block__content inner"
								style="padding: 10px 0 18px 0; font-size: 16px; line-height: 27px;" align="left">
								Read the
								<a href="https://folktells.com/our-story/" target="_blank">story behind Folktells</a>, and
								find
								<a target="_blank" href="https://folktells.com/faq/">FAQs</a>, <a target="_blank"
								  href="https://folktells.com/learn/">Learning
								  Articles</a> &amp; <a target="_blank" href="https://folktells.com/get-in-touch/">Support</a>
								on our <a target="_blank" href="https://folktells.com">website</a>.
							  </td>
							</tr>
						  </table>
						</td>
					  </tr>
					</table>
					<table class="container" border="0" cellpadding="0" cellspacing="0" width="100%"
					  style="padding-top: 25px;" align="center">
					  <tr>
						<td align="center">
						  <table class="container" border="0" cellpadding="0" cellspacing="0" width="620" align="center"
							style="border-bottom: solid 1px #eeeeee; width: 620px;">
							<tr>
							  <td align="center">&nbsp;</td>
							</tr>
						  </table>
						</td>
					  </tr>
					</table>
					<table class="container" border="0" cellpadding="0" cellspacing="0" width="100%">
					  <tr>
						<td align="center" valign="top">
						  <table class="container" border="0" cellpadding="0" cellspacing="0" width="620"
							style="width: 620px;">
							<tr>
							  <td class="cta-block__title inner"
								style="padding: 35px 0 0 0; font-size: 26px; text-align: center;">
								About Us</td>
							</tr>
							<tr>
							  <td class="cta-block__content inner"
								style="color:#222222;padding: 20px 0 27px 0; font-size: 16px; line-height: 27px; text-align: center;">
								Folktells is a mobile app designed for seniors and their families &amp; friends. Because they
								tend
								not to be as tech-savvy, seniors often feel
								left out. Folktells delivers the benefits of a connected world,
								simply,
								without the confusion.</td>
							</tr>
						  </table>
						  <table cellpadding="0" cellspacing="20" width="100%">
							<tr>
							  <td style="width: 50%;" width="50%" valign="top" align="center"><a target="_blank"
								  href="https://itunes.apple.com/ca/app/folktells/id1489217069?mt=8">
								  <img alt="Download on the App Store" style="width:100%; max-width:135px;" width="135"
									src="https://folktells.com/wp-content/uploads/2020/03/appstore-badge@2x.png" />
								</a>
							  </td>
							  <td style="width: 50%;" width="50%" valign="top" align="center"><a target="_blank"
								  href="https://play.google.com/store/apps/details?id=com.csdcorp.app.folktells.android">
								  <img alt="Get it on Google Play" style="width:100%; max-width:135px;" width="135"
									src="https://folktells.com/wp-content/uploads/2020/03/google-play-badge@2x.png" /></a>
							  </td>
							</tr>
						  </table>
						</td>
					  </tr>
					</table>
					<table class="container" border="0" cellpadding="0" cellspacing="0" width="100%"
					  style="padding-top: 25px;" align="center">
					  <tr>
						<td align="center">
						  <table class="container" border="0" cellpadding="0" cellspacing="0" width="620" align="center"
							style="border-bottom: solid 1px #eeeeee; width: 620px;">
							<tr>
							  <td align="center">&nbsp;</td>
							</tr>
						  </table>
						</td>
					  </tr>
					</table>
					<table class="container" border="0" cellpadding="0" cellspacing="0" width="100%" align="center">
					  <tr>
						<td align="center">
						  <table class="container" border="0" cellpadding="0" cellspacing="0" width="620" align="center">
							<tr>
							  <td class="inner" style="text-align: center; padding: 50px 0 10px 0;">
								<a href="https://folktells.com"
								  style="font-size: 28px; text-decoration: none; color: #3e4c84;">Folktells</a>
							  </td>
							</tr>
							<tr>
							  <td align="center">
								<table width="60" border="0" cellpadding="0" cellspacing="0"
								  style="width: 60px; height: 60px;">
								  <tr>
									<td align="center" width="60" height="60" style="width: 60px; height: 60px;">
									  <img width="60" height="60" style="width: 60px; height: 60px;" alt="Folktells App Icon"
										src="https://folktells.com/wp-content/uploads/2020/02/folktells-icon.png" /></td>
								  </tr>
								</table>
							  </td>
							</tr>
							<tr>
							  <td class="inner"
								style="color: #6a6a6a;text-align: center; font-size: 15px; padding: 10px 0 60px 0; line-height: 22px;">
								Copyright &copy; 2020 <a href="http://csdcorp.com/" target="_blank"
								  style="text-decoration: none; border-bottom: 1px solid #d5d5d5; color: #6a6a6a;">Corner
								  Software Development Corp.</a><br />All rights reserved.</td>
							</tr>
						  </table>
						</td>
					  </tr>
					</table>
				  </td>
				</tr>
			  </table>
	  
			</td>
		  </tr>
		</table>
	  </body>
	  
	  </html>	`,
		TextBody: `
		Welcome to Folktells!

		We are excited to have you join the Folktells community.
		
		Getting Started
		
		Visit our web site to learn:
		- how to create & share stories: https://folktells.com/how-to/create-and-manage-stories/
		- how to connect to shared albums from Google Photos and iCloud: https://folktells.com/how-to/albums-from-google-photos/
		- how to view events from your calendars: https://folktells.com/how-to/select-calendars-to-display/
		and more.
		
		
		Invite your friends and family to join Folktells and share stories & photos about the special events, people & pets, and adventures that life brings your way.
		* from the menu, choose Sharing
		* on the Groups tab, be sure you are signed in and that Sharing is ON
		* tap New and enter the name of your group
		* tap create
		* expand your new group section
		* invite your family & friends to join.
		
		Find out more on Folktells.com: https://folktells.com
		Check out:
		* Our FAQ: https://folktells.com/faq/
		* Learning Articles: https://folktells.com/learn/
		And if you need to, get in touch: https://folktells.com/get-in-touch/
		
		About Us
		Folktells is a mobile app designed for seniors and their families & friends. Seniors and others who find technology
		challenging often get left out and feel disconnected from their social-app-using family members. Folktells delivers
		the benefits of a connected world, simply, without the confusion.
		
		Copyright (c) 2020 Corner Software Development Corp.
		https://csdcorp.com/
		All rights reserved.		`,
	}
}
