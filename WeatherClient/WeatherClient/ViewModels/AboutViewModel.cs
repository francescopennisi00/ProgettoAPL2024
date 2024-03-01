using CommunityToolkit.Mvvm.Input;
using System.Windows.Input;

namespace WeatherClient.ViewModels;

internal class AboutViewModel
{
    public string Title => AppInfo.Name;
    public string Version => AppInfo.VersionString;
    public string MoreInfoUrl => "https://github.com/francescopennisi00/ProgettoAPL2024/";
    public string Message => "This app is written in XAML and C# with .NET MAUI. It is the client side of a larger application whose server is written in Python and Go.";
    public ICommand ShowMoreInfoCommand { get; }

    public AboutViewModel()
    {
        ShowMoreInfoCommand = new AsyncRelayCommand(ShowMoreInfo);
    }

    async Task ShowMoreInfo() =>
        await Launcher.Default.OpenAsync(MoreInfoUrl);
}
