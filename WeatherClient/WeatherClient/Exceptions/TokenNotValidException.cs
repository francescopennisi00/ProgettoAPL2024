namespace WeatherClient.Exceptions
{
    internal class TokenNotValidException : Exception
    {
        private string message;
        public TokenNotValidException(string message)
        {
            this.message = message;
        }
    }
}
