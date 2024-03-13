using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace WeatherClient.Utilities
{
    internal class Constants
    {
        public const string urlUpdate = "http://weather.com:8080/wms/update_rules";
        public const string urlDelete = "http://weather.com:8080/wms/update_rules/delete_user_constraints_by_location";
        public const string urlShow = "http://weather.com:8080/wms/show_rules";
        public const string urlSignup = "http://weather.com:8080/usermanager/register";
        public const string urlLogin = "http://weather.com:8080/usermanager/login";
        public const string urlDeletAccount = "http://weather.com:8080/usermanager/delete_account";
        public static string tokenPath = FileSystem.AppDataDirectory + @"\JWT_token.txt";
    }
}
