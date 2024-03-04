namespace WeatherClient.Exceptions
{
    internal class ServerException : Exception
    {
        public string Errormessage { get; private set; }
        public ServerException(string message)
        {
            this.Errormessage = message;
        }
    }
}
