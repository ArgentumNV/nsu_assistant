using Telegram.Bot;
using Telegram.Bot.Types;
using Telegram.Bot.Types.Enums;

using Microsoft.Extensions.Configuration;


// загрузка конфигурации из User Secrets, ссылаясь на класс Program
var config = new ConfigurationBuilder().AddUserSecrets<Program>().Build();

// получаем токен
var token = config["BotToken"];
if(string.IsNullOrEmpty(token))
{
    Console.WriteLine("Ошибка: токен не найден. Проверь, что ты выполнил 'dotnet user-secrets set \"BotToken\" \"...\"'");
    return;
}



using var cts = new CancellationTokenSource();

var bot = new TelegramBotClient(token, cancellationToken: cts.Token);
var me =  await bot.GetMe();
bot.OnMessage += OnMessage;

//Console.WriteLine($"EMAE! i am user {me.Id}, name is {me.FirstName}");

// заставляет программу ждать, пока не нажат Enter
Console.WriteLine($"@{me.Username} is running... Press Enter to exit");
Console.ReadLine();
cts.Cancel(); // stop the bot 

async Task OnMessage(Message msg, UpdateType type)
{
    if (msg.Text is null) return;

    Console.WriteLine($"Received {type} '{msg.Type}' in {msg.Chat}: {msg.Text}");

    //echo
    await bot.SendMessage(msg.Chat, $"Бляя какой же ты ебень... Что за '{msg.Text}'?");
}



// явное объявление класса Program (нужен для AddUserSecrets<Program>)
public partial class Program {}