namespace WeatherClient.Exceptions
{
    internal class BadRequestException : Exception
    {
        public string Errormessage { get; private set; }
        public BadRequestException(string message)
        {
            Errormessage = message;
        }
    }
}
