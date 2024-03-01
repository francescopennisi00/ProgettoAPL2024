using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace WeatherClient.Utilities
{
    internal class Constants
    {
        public const string urlUpdate = "http://weather.com:8080/update_rules";
        public const string urlDelete = "http://weather.com:8080/update_rules/delete_user_constraints_by_location";
        public const string urlShow = "http://weather.com:8080/show_rules";
    }
}
