import argparse

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
            
            first_li = driver.find_element(By.CSS_SELECTOR, "ul.serversList li:first-child")

            data_link = first_li.get_attribute("data-link")

            ep_links.append(data_link)
            
    finally:
        if driver:
            CloseDriver()

    return ep_links



parser = argparse.ArgumentParser("egydead_browser_fetcher")
parser.add_argument("links", help="list of egydead episode links", type=str)
args = parser.parse_args()
links = get_episodes_servers(urls=args.links.split(","))
print(links)
