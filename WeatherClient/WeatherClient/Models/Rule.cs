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

    private static HttpResponseMessage DoHttpRequest(string url, string content)
    {
        string token = TokenUtility.GetToken();
        if (token == "null")
        {
            // user is not logged in
            return new HttpResponseMessage(HttpStatusCode.Unauthorized);
        }
        using (HttpClient httpC = new HttpClient())
        {
            httpC.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", token);
            if (content != null)
            {
                HttpRequestMessage request = new HttpRequestMessage
                {
                    Method = HttpMethod.Post,
                    RequestUri = new Uri(url),
                    Content = new StringContent(content, Encoding.UTF8, "application/json")
                };
                return httpC.Send(request);
            }
            else
            {
                HttpRequestMessage request = new HttpRequestMessage
                {
                    Method = HttpMethod.Get,
                    RequestUri = new Uri(url),
                    Content = null
                };
                return httpC.Send(request);
            }
        }
    }

    public async Task<string> Save()
    {
        string jsonData = JsonUtility.SerializeJSON(this);
        var response = DoHttpRequest(Constants.urlUpdate, jsonData);
        if (response.StatusCode == HttpStatusCode.Unauthorized)
        {
            throw new TokenNotValidException("JWT Token provided is not valid. Login required.");
        }
        if (response.StatusCode == HttpStatusCode.BadRequest)
        {
            throw new BadRequestException("Bad request! Please enter again.");
        }
        else if (response.StatusCode != HttpStatusCode.OK)
        {
            throw new ServerException("Failed to load rules due an internal server error.");
        }
        var stringReturned = await response.Content.ReadAsStringAsync();
        // next instructions only if request is an update and this update is an insert of a new rule
        if (Id == String.Empty)
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
        }
        else
        {
            return stringReturned;
        }
    }

    public void Delete()
    {
        string jsonData = JsonUtility.SerializeJSON(this);
        var response = DoHttpRequest(Constants.urlDelete, jsonData);
        if (response.StatusCode == HttpStatusCode.Unauthorized)
        {
            throw new TokenNotValidException("JWT Token provided is not valid. Login required.");
        }
        if (response.StatusCode == HttpStatusCode.BadRequest)
        {
            throw new BadRequestException("Bad request! Please enter again.");
        }
        else if (response.StatusCode != HttpStatusCode.OK)
        {
            throw new ServerException("Failed to load rules due an internal server error.");
        }
        return;
    }

    public static async Task<Rule> Load(string id)
    {
        try
        {
            IEnumerable<Rule> rules = await LoadAll();
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
        catch (TokenNotValidException)
        {
            throw new TokenNotValidException("JWT Token provided is not valid. Login required.");
        } 
        catch (ServerException)
        {
            throw new ServerException("Failed to load rules due an internal server error.");
        }
    }

    public static async Task<IEnumerable<Rule>> LoadAll()
    {
        var response = DoHttpRequest(Constants.urlShow, null);

        if (response.IsSuccessStatusCode)
        {
            string responseContent = await response.Content.ReadAsStringAsync();
            IEnumerable<Rule> rules = JsonUtility.DeserializeJSON<IEnumerable<Rule>>(responseContent);
            if (rules == null)
            {
                return new List<Rule>();
            }
            return rules;
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
}
