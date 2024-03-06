namespace WeatherClient.Exceptions
{
    internal class TokenNotValidException : Exception
    {
        public string Errormessage { get; private set; }
        public TokenNotValidException(string message)
        {
            Errormessage = message;
        }
    }
}
