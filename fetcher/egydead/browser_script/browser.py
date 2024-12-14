import argparse
from urllib.parse import urlparse
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC



driver = None

def Initialize():
    global driver
    options = webdriver.ChromeOptions()
    options.add_argument("--incognito") 
    options.add_argument('--ignore-ssl-errors=yes')
    options.add_argument('--ignore-certificate-errors')
    options.add_argument('--disable-dev-shm-usage')
    options.add_argument("--headless")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-gpu")
    options.add_argument("--disable-extensions")
    options.page_load_strategy = 'none'
    prefs = {
        "profile.managed_default_content_settings.images": 2,
        "profile.managed_default_content_settings.css": 2,
    }
    options.add_experimental_option("prefs", prefs)
    driver = webdriver.Chrome(options=options)
    driver.implicitly_wait(5)
    return driver
    
def CloseDriver():
    global driver
    driver.quit()

def get_episodes_servers(urls):
    ep_links = []
    try:
        Initialize()
        for url in urls:
            driver.get(url)
            
            watch_btn = driver.find_element(By.CSS_SELECTOR, ".watchNow button")
            watch_btn.click()
            WebDriverWait(driver, 30).until(
                EC.presence_of_element_located((By.CSS_SELECTOR, ".watchAreaMaster"))
            )
            driver.execute_script("window.stop();")
            first_li = driver.find_element(By.CSS_SELECTOR, "ul.serversList li:first-child")
            data_link = first_li.get_attribute("data-link")
            ep_links.append(data_link)
            
    finally:
        if driver:
            CloseDriver()

    return ep_links


def search(url, q):
    try:
        Initialize()
        parsed_url = urlparse(url)
        base_url = f"{parsed_url.scheme}://{parsed_url.netloc}"
        driver.get(base_url)
        search_input = driver.find_element(By.NAME, "s")
        search_input.send_keys(q)
        found_search_results = None
        def element_displayed_block(driver):
            element = driver.find_element(By.CSS_SELECTOR, ".searchLive")
            found_search_results = element
            return element.value_of_css_property('display') == 'block'
        WebDriverWait(driver, 10).until(element_displayed_block)
        driver.execute_script("window.stop();")
        search_result = driver.find_element(By.CSS_SELECTOR, ".searchLive")
        return search_result.get_attribute("innerHTML")

    finally:
        if driver:
            CloseDriver()

parser = argparse.ArgumentParser("egydead_browser_fetcher")
parser.add_argument('-l', '--links',help="list of egydead episode links", type=str)
parser.add_argument('-c', '--command',help="command type", choices=['search', 'episode_servers'])
parser.add_argument('-su', '--search_url',help="the url which will be used for searching", type=str)
parser.add_argument('-q', '--search',help="search query", type=str)

args = parser.parse_args()

if args.command == "episode_servers":
    links = get_episodes_servers(urls=args.links.split(","))
    print(links)
elif args.command == "search":
    html_search_result = search(args.search_url, args.search)
    print(html_search_result)
