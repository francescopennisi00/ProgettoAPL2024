namespace WeatherClient.Exceptions
{
    internal class ServerException : Exception
    {
        private string message;
        public ServerException(string message)
        {
            this.message = message;
        }
    }
}
