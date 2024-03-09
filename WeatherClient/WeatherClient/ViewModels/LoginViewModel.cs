using CommunityToolkit.Mvvm.ComponentModel;
using System.Windows.Input;
using WeatherClient.Exceptions;
using WeatherClient.Models;
using WeatherClient.Views;
namespace WeatherClient.ViewModels;

internal class LoginViewModel : ObservableObject, IQueryAttributable
{
    private User _user;

    public bool IsVisibleLogin { get; set; } = true;
    public bool IsVisibleLogout { get; set; } = false;
    public bool IsVisiblePasswordDeleteAccount { get; set; } = false;

    public string UserName
    {
        get => _user.UserName;
        set
        {
            _user.UserName = value;
            OnPropertyChanged(nameof(UserName));
        }
    }
    public string Password
    {
        get => _user.Password;
        set 
        { 
            _user.Password = value; 
            OnPropertyChanged(nameof(Password));
        }
    }

    public ICommand Login { get; set; }
    public ICommand Signup { get; set; }
    public ICommand Logout { get; set; }
    public ICommand DeleteAccount { get; set; }

    public LoginViewModel()
    {
        _user = new User();

        // we assume that user is logged in, if he is not, then the other constructor is called
        IsVisibleLogin = false;  
        IsVisibleLogout = true;
        var email = Utilities.TokenUtility.findUsernameByToken();
        if (email != "null")
        {
            UserName = email;
        } else
        {
            // user is not logged in and we change visibility of UI elements
            IsVisibleLogin = true;
            IsVisibleLogout = false;
        }

        Login = new Command(LoginClicked);
        Signup = new Command(SignupClicked);
        Logout = new Command(LogoutClicked);
        DeleteAccount = new Command(DeleteAccountClicked);
    }

    private async void LoginClicked()
    {
        try
        {
            await _user.Login();
            IsVisibleLogin = false;
            IsVisibleLogout = true;
            OnPropertyChanged(nameof(IsVisibleLogin));
            OnPropertyChanged(nameof(IsVisibleLogout));
            // if login was successfull we set credentials and go to Your Rules page
            await Shell.Current.GoToAsync("//AllRulesRoute");
        }
        catch (UsernamePswWrongException exc)
        {
            var title = "Wrong Authentication!";
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
            var title = "Internal Error!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
    }

    private async void SignupClicked()
    {
        SignupPage page = new SignupPage();
        if (!string.IsNullOrEmpty(UserName)) // if the user has already entered the username during login, we facilitate the signup process by automatically inserting it into the new page
            page.BindingContext = new SignupViewModel(UserName);
        await App.Current.MainPage.Navigation.PushAsync(page);
    }

    private async void LogoutClicked()
    {
        try
        {
            _user.Logout();
            IsVisibleLogin = true;
            IsVisibleLogout = false;
            OnPropertyChanged(nameof(IsVisibleLogin));
            OnPropertyChanged(nameof(IsVisibleLogout));
            _user.Password = String.Empty;
            _user.UserName = String.Empty;
            OnPropertyChanged(nameof(Password));
            OnPropertyChanged(nameof(UserName));
        }
        catch (Exception exc)
        {
            var title = "Error!";
            var message = exc.Message;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
    }
 
    private async void DeleteAccountClicked()
    {
        try
        {
            IsVisiblePasswordDeleteAccount = true;
            OnPropertyChanged(nameof(IsVisiblePasswordDeleteAccount));
            if (Password != null)  // instead, if Password is null then user will have to insert it in order to delete account
            {
                _user.DeleteAccount();
                IsVisibleLogin = true;
                IsVisibleLogout = false;
                IsVisiblePasswordDeleteAccount = false;
                OnPropertyChanged(nameof(IsVisibleLogin));
                OnPropertyChanged(nameof(IsVisibleLogout));
                OnPropertyChanged(nameof(IsVisiblePasswordDeleteAccount));
                _user.Password = String.Empty;
                _user.UserName = String.Empty;
                OnPropertyChanged(nameof(Password));
                OnPropertyChanged(nameof(UserName));
            }
        }
        catch (Exception exc)
        {
            _user.Password = String.Empty;  // we reset Password at null in order to allow user to reinsert password
            var title = "Error!";
            var message = exc.Message;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
    }

    void IQueryAttributable.ApplyQueryAttributes(IDictionary<string, object> query)
    {
        if (query.ContainsKey("registered"))
        {
            IsVisibleLogin = false;
            IsVisibleLogout = true;
            OnPropertyChanged(nameof(IsVisibleLogin));
            OnPropertyChanged(nameof(IsVisibleLogout));
            var username = query["registered"].ToString();
            if (!string.IsNullOrEmpty(username))
            {
                UserName = username;
            }
        }
    }
}
