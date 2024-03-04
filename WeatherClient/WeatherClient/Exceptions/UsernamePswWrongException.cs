using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace WeatherClient.Exceptions
{
    internal class UsernamePswWrongException : Exception
    {
        public string Errormessage { get; private set; }
        public UsernamePswWrongException(string message)
        {
            this.Errormessage = message;
        }

    }
}
