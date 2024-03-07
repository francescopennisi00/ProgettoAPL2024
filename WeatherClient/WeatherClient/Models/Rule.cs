using System.Text;
using System.Net;
using Newtonsoft.Json;
using WeatherClient.Utilities;
using WeatherClient.Exceptions;
using System.Text.RegularExpressions;

namespace WeatherClient.Models;

internal class Rule
{
    [JsonProperty("rules")]
    public WeatherParameters? Rules { get; set; }

    [JsonProperty("trigger_period")]
    public string? TriggerPeriod { get; set; }

    [JsonProperty("location")]
    public List<string>? Location { get; set; }

    [JsonProperty("id")]
    public string? Id { get; set; }


    public Rule()
    {
        Rules = new WeatherParameters();
        TriggerPeriod = string.Empty;
        Location = null;
        Id = string.Empty;
    }

    public Rule(string id, string locName, string lat, string lon, string cCode, string sCode, string triggerP, string maxT,
        string minT, string maxH, string minH, string maxP, string minP, string maxWS, string minWS, string direction, string rain,
        string snow, string maxC, string minC)
    {
        Rules = new WeatherParameters(maxT, minT, maxH, minH, maxP, minP, maxWS, minWS, direction, rain, snow, maxC, minC);
        TriggerPeriod = triggerP;
        Location = new List<string> { locName, lat, lon, cCode, sCode };
        Id = id;
    }

    private static string GetToken()
    {
        // get the folder where the tokes is stored
        string appDataPath = FileSystem.AppDataDirectory + @"\JWT_token.txt";
        string token = File.ReadAllText(appDataPath);
        string token_to_return = token.Replace("\n", "").Replace("\r", "");
        return token_to_return;
    }

    private async Task<string> DoRequest(string url)
    {
        string token = GetToken();
        using (HttpClient httpC = new HttpClient())
        {
            httpC.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", token);
            string jsonData = JsonUtility.SerializeJSON(this);
            HttpRequestMessage request = new HttpRequestMessage
            {
                Method = HttpMethod.Post,
                RequestUri = new Uri(url),
                Content = new StringContent(jsonData, Encoding.UTF8, "application/json")
            };
            var prova = request.Content;
            HttpResponseMessage response = httpC.Send(request);
            if ((int)response.StatusCode == 401)
            {
                throw new TokenNotValidException("JWT Token provided is not valid. Login required.");
            }
            if ((int)response.StatusCode == 400)
            {
                throw new BadRequestException("Bad request! Please enter again.");
            }
            else if ((int)response.StatusCode != 200)
            {
                throw new ServerException("Failed to load rules due an internal server error.");
            }
            var stringReturned = await response.Content.ReadAsStringAsync();
            // next instructions only if request is an update, not a delete and this update is an insert of a new rule
            if (url == Constants.urlUpdate && Id == String.Empty)
            {
                // extracting id of the new rule inserted by regular expressions
                Regex regex = new Regex(@"\d+");
                Match match = regex.Match(stringReturned);
                if (match.Success)
                {
                    return match.Value;
                }
                else
                {
                    throw new ServerException("Error: server has not returned id of the rule inserted.");
                }
            } else
            {
                return stringReturned;
            }
        }
    }

    public async Task<string> Save()
    {
        return await DoRequest(Constants.urlUpdate);
    }

    public async void Delete()
    {
        await DoRequest(Constants.urlDelete);
    }

    public static Rule Load(string id)
    {
        IEnumerable<Rule> rules = LoadAll().Result;
        foreach (Rule rule in rules)
        {
            if (rule.Id == id)
            {
                return new()
                {
                    Rules = rule.Rules,
                    Id = rule.Id,
                    TriggerPeriod = rule.TriggerPeriod,
                    Location = rule.Location
                };
            }
        }
        throw new Exception("Error! Rule not found!");
    }
 

    public static async Task<IEnumerable<Rule>> LoadAllAsync()
    {
        try
        {
            string url = Constants.urlShow;
            string token = GetToken();
            HttpClient httpC = new HttpClient();
            httpC.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", token);
            HttpRequestMessage request = new HttpRequestMessage
            {
                Method = HttpMethod.Get,
                RequestUri = new Uri(url),
                Content = null
            };
            HttpResponseMessage response = httpC.Send(request);

            if (response.IsSuccessStatusCode)
            {
                string responseContent = await response.Content.ReadAsStringAsync();
                return JsonUtility.DeserializeJSON<IEnumerable<Rule>>(responseContent);
            }
            else if (response.StatusCode == HttpStatusCode.Unauthorized)
            {
                throw new TokenNotValidException("JWT Token provided is not valid. Login required.");
            }
            else
            {
                throw new ServerException("Failed to load rules due an internal server error.");
            }
        }
        catch (HttpRequestException)
        {
            throw new ServerException("Failed to load rules due an internal server error.");
        }
    }

    public static async Task<IEnumerable<Rule>> LoadAll()
    {
        try
        {
            IEnumerable<Rule> rules = await LoadAllAsync();
            if (rules == null)
            {
                return new List<Rule>();
            }
            return rules;
        }
        catch(TokenNotValidException)
        {
            throw new TokenNotValidException("JWT Token provided is not valid. Login required.");
        }
        catch(ServerException)
        {
            throw new ServerException("Failed to load rules due an internal server error.");
        }
    }
    /*
    private async Task<HttpResponseMessage> DoRequest(HttpMethod httpreq, string url, string json)
    {
        try
        {
            HttpRequestMessage request = new HttpRequestMessage
            {
                Method = httpVerb,
                RequestUri = new Uri(url),
                Content = null
            };
            if (!string.IsNullOrEmpty(UserService.Instance.Token))
                richiesta.Headers.Add("Authorization", UserService.Instance.Token); //aggiungo l'eventuale token, se disponibile
            if (!string.IsNullOrEmpty(json))
                richiesta.Content = new StringContent(json, Encoding.UTF8, "application/json"); //aggiungo l'eventuale contenuto

            HttpResponseMessage risposta = await _client.SendAsync(richiesta);

            return risposta;
        }
        catch (Exception e)
        {
            return new HttpResponseMessage(HttpStatusCode.BadRequest);
        }
    }
    */
}