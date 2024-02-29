using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace WeatherClient.Models;

internal class Rule
{
    public string Filename { get; set; }
    public string Text { get; set; }
    public DateTime Date { get; set; }
}
