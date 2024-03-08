using Newtonsoft.Json;
using System.Net;
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
            if (response.StatusCode == HttpStatusCode.Unauthorized)
            {
                throw new UsernamePswWrongException("Email or password wrong. Retry!");
            }
            if (response.StatusCode == HttpStatusCode.BadRequest)
            {
                throw new BadRequestException("Bad request! Please enter again.");
            }
            else if (response.StatusCode != HttpStatusCode.OK)
            {
                throw new ServerException("Failed to login due an internal server error.");
            }
            string responseString = await response.Content.ReadAsStringAsync();
            var token = Utilities.TokenUtility.ExtractToken(responseString);
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
            if (response.StatusCode == HttpStatusCode.BadRequest)
            {
                throw new EmailAlreadyInUseException("Email already in use. Try to sign in!");
            }
            else if (response.StatusCode != HttpStatusCode.OK)
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
            if (response.StatusCode == HttpStatusCode.Unauthorized)
            {
                throw new UsernamePswWrongException("Email or password wrong. Retry!");
            }
            if (response.StatusCode == HttpStatusCode.BadRequest)
            {
                throw new BadRequestException("Bad request! Please enter again.");
            }
            else if (response.StatusCode != HttpStatusCode.OK)
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
