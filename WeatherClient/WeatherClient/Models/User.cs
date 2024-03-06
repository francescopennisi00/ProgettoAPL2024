using Newtonsoft.Json;
using System.Text;
using WeatherClient.Exceptions;

namespace WeatherClient.Models;

internal class User
{
    [JsonProperty("email")]
    public string? UserName { get; set; }
    [JsonProperty("password")]
    public string? Password { get; set; }

    public User(string userName, string password)
    {
        UserName = userName;
        Password = password;
    }

    public User()
    {
    }

    private static string ExtractToken(string response)
    {

        int colonIndex = response.IndexOf(':');
        string token = response.Substring(colonIndex + 2);
        return token;
    }

    public async Task Login()
    {
        using (HttpClient httpC = new HttpClient())
        {
            string jsonData = Utilities.JsonUtility.SerializeJSON(this);
            HttpRequestMessage request = new HttpRequestMessage
            {
                Method = HttpMethod.Post,
                RequestUri = new Uri(Utilities.Constants.urlLogin),
                Content = new StringContent(jsonData, Encoding.UTF8, "application/json")
            };
            HttpResponseMessage response = httpC.Send(request);
            if ((int)response.StatusCode == 401)
            {
                throw new UsernamePswWrongException("Username or password wrong. Retry!");
            }
            if ((int)response.StatusCode == 400)
            {
                throw new BadRequestException("Bad request! Please enter again.");
            }
            else if ((int)response.StatusCode != 200)
            {
                throw new ServerException("Failed to login due an internal server error.");
            }
            string responseString = await response.Content.ReadAsStringAsync();
            var token = ExtractToken(responseString);
            File.WriteAllText(Utilities.Constants.tokenPath, token);
        }
    }

    public void SignUp()
    {
        using (HttpClient httpC = new HttpClient())
        {
            string jsonData = Utilities.JsonUtility.SerializeJSON(this);
            HttpRequestMessage request = new HttpRequestMessage
            {
                Method = HttpMethod.Post,
                RequestUri = new Uri(Utilities.Constants.urlSignup),
                Content = new StringContent(jsonData, Encoding.UTF8, "application/json")
            };
            HttpResponseMessage response = httpC.Send(request);
            if ((int)response.StatusCode == 401)
            {
                throw new EmailAlreadyInUseException("Email already in use. Try to sign in!");
            }
            if ((int)response.StatusCode == 400)
            {
                throw new BadRequestException("Bad request! Please enter again.");
            }
            else if ((int)response.StatusCode != 200)
            {
                throw new ServerException("Failed to signup due an internal server error.");
            }
        }
    }

    public void Logout()
    {

        if (File.Exists(Utilities.Constants.tokenPath))
        {
            // delete JWT Token file
            File.Delete(Utilities.Constants.tokenPath);
        }
        else
        {
            throw new FileNotFoundException("JWT Token file not found.");
        }
        
    }

    public void DeleteAccount()
    {
        using (HttpClient httpC = new HttpClient())
        {
            string jsonData = Utilities.JsonUtility.SerializeJSON(this);
            HttpRequestMessage request = new HttpRequestMessage
            {
                Method = HttpMethod.Post,
                RequestUri = new Uri(Utilities.Constants.urlDeletAccount),
                Content = new StringContent(jsonData, Encoding.UTF8, "application/json")
            };
            HttpResponseMessage response = httpC.Send(request);
            if ((int)response.StatusCode == 401)
            {
                throw new UsernamePswWrongException("Username or password wrong. Retry!");
            }
            if ((int)response.StatusCode == 400)
            {
                throw new BadRequestException("Bad request! Please enter again.");
            }
            else if ((int)response.StatusCode != 200)
            {
                throw new ServerException("Failed to delete account due an internal server error.");
            }
            try
            {
                Logout();
            }
            catch (FileNotFoundException)
            {
                throw new FileNotFoundException("JWT Token file not found.");
            }
        }
    }
}
