using Newtonsoft.Json.Linq;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace WeatherClient.Utilities
{
    internal class TokenUtility
    {
        public static string GetToken()
        {
            // get the folder where the token is stored
            string appDataPath = FileSystem.AppDataDirectory + @"\JWT_token.txt";
            if (File.Exists(appDataPath))
            {
                string token = File.ReadAllText(appDataPath);
                // remove carriage return and new line characters from token writed in the text file
                string token_to_return = token.Replace("\n", "").Replace("\r", "");
                return token_to_return;
            }
            else
            {
                return "null";
            }

        }

        public static string ExtractToken(string response)
        {

            int colonIndex = response.IndexOf(':');
            string token = response.Substring(colonIndex + 2);
            return token;
        }

        public static string findUsernameByToken()
        {
            var token = GetToken();
            if (token == "null")
            {
                return token;
            }
            // Decode JWT Token
            var tokenHandler = new JwtSecurityTokenHandler();
            var jwtSecurityToken = tokenHandler.ReadToken(token) as JwtSecurityToken;

            if (jwtSecurityToken != null)
            {
                // Obtain "email" claim
                string email = jwtSecurityToken.Claims.FirstOrDefault(claim => claim.Type == "email")?.Value;
                if (email != null)
                {
                    return email;
                } else
                {
                    return "null";
                }
            }
            else
            {
                return "null";
            }
        }
    }
}
