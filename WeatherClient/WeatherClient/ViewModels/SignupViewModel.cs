using System.Windows.Input;
using WeatherClient.Exceptions;
using System.Text.RegularExpressions;

namespace WeatherClient.ViewModels;

internal class SignupViewModel
{

    private Models.User _user;

    public string? UserName
    {
        get => _user.UserName;
        set => _user.UserName = value;
    }
    public string? Password
    {
        get => _user.Password;
        set => _user.Password = value;
    }
    public string? ConfirmPassword { get; set; }

    public ICommand Signup { get; set; }

    public SignupViewModel(string username) : this()
    {
        UserName = username;
    }

    public SignupViewModel()
    {
        _user = new Models.User();
        Signup = new Command(SignupClicked);
    }

    private async void SignupClicked()
    {
        // data validation checks
        if (string.IsNullOrEmpty(UserName) || string.IsNullOrEmpty(Password))
        {
            await App.Current.MainPage.DisplayAlert("Warning", "Fill in the email and password fields.", "OK");
            return;
        }
        if (!Regex.IsMatch(UserName, @"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"))
        {
            await App.Current.MainPage.DisplayAlert("Warning", "Email not valid.", "OK");
            return;
        }
        if (ConfirmPassword != Password)
        {
            await App.Current.MainPage.DisplayAlert("Warning!", "The password and confirm password fields must match.", "OK");
            return;
        }

        try
        {
            _user.SignUp();
            await _user.Login();
            await Shell.Current.GoToAsync($"..?registered={UserName}");
            await Shell.Current.GoToAsync($"//AllRulesRoute?login={true}");
        }
        catch (EmailAlreadyInUseException exc)
        {
            var title = "Error!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
        catch (BadRequestException exc)
        { 
            var title = "Warning!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
        catch (ServerException exc)
        {
            var title = "Warning!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
    }

}
