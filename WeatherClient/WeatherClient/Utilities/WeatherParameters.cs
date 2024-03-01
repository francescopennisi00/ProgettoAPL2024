using Newtonsoft.Json;

namespace WeatherClient.Utilities;

internal class WeatherParameters
    {
		[JsonProperty("max_temp")]
		public string? MaxTemp { get; set; }

		[JsonProperty("min_temp")]
		public string? MinTemp { get; set; }

		[JsonProperty("max_humidity")]
		public string? MaxHumidity { get; set; }

		[JsonProperty("min_humidity")]
		public string? MinHumidity { get; set; }

		[JsonProperty("max_pressure")]
		public string? MaxPressure { get; set; }

		[JsonProperty("min_pressure")]
		public string? MinPressure { get; set; }

		[JsonProperty("max_wind_speed")]
		public string? MaxWindSpeed { get; set; }

		[JsonProperty("min_wind_pressure")]
		public string? MinWindSpeed { get; set; }

		[JsonProperty("wind_direction")]
		public string? WindDirection { get; set; }

		[JsonProperty("rain")]
		public string? Rain { get; set; }

		[JsonProperty("snow")]
		public string? Snow { get; set; }

		[JsonProperty("max_cloud")]
		public string? MaxCloud { get; set; }

		[JsonProperty("min_cloud")]
		public string? MinCloud { get; set; }

	public WeatherParameters()
    {
		MaxTemp = null;
		MinTemp = null;
		MaxHumidity = null;
		MinHumidity = null;
		MaxPressure = null;
		MinPressure = null;
		MaxWindSpeed = null;
		MinWindSpeed = null;
		WindDirection = null;
		Rain = null;
		Snow = null;
		MaxCloud = null;
		MinCloud = null;
	}

    public WeatherParameters(string maxT, string minT, string maxH, string minH, string maxP, string minP, string maxWS, string minWS,
		string direction, string rain, string snow, string maxC, string minC)
    {
        MaxTemp = maxT;
        MinTemp = minT;
        MaxHumidity = maxH;
        MinHumidity = minH;
        MaxPressure = maxP;
        MinPressure = minP;
        MaxWindSpeed = maxWS;
        MinWindSpeed = minWS;
        WindDirection = direction;
        Rain = rain;
        Snow = snow;
        MaxCloud = maxC;
        MinCloud = minC;
    }

}
