require 'mechanize'
require 'nokogiri'
require 'csv'
require 'json'

class Scraper

  def initialize
  @vc_username = ENV['VERACROSS_USERNAME']
  @vc_password = ENV['VERACROSS_PASSWORD']
  @vc_client = ENV['VERACROSS_CLIENT']
  @query_id = ARGV[0]
  @url = "https://esweb.veracross.com/#{@vc_client}/esweb.asp?WCI=Results&Query=#{@query_id}"
  end

  def is_int(str)
    # Check if a string should be an integer
    return !!(str =~ /^[-+]?[1-9]([0-9]*)?$/)
  end

  def login
    agent = Mechanize.new
    # agent.log = @logger #enable mechanize logging
    agent.agent.http.verify_mode = OpenSSL::SSL::VERIFY_NONE
    page = agent.get(@url)
    form = page.forms.first
    username = form.fields.first
    password = form.fields.last
    username.value = @vc_username
    password.value = @vc_password
    button = form.buttons.last
    logged_in_page = form.submit(button)
    doc = logged_in_page
    return doc,agent
  end

  def export_csv
    doc,agent = login
    link = doc.link_with(text: ' Export')
    query_page = link.click
    link = query_page.link_with(text: 'Comma-Delimited Text File')
    csv_page = link.click
    link = csv_page.links.last
    csv_output = agent.get(link.attributes.first[1]).body
    return csv_output
  end

  def print_json
    lines = CSV.parse(export_csv)
    keys = lines.delete lines.first
    data = lines.map do |values|
      is_int(values) ? values.to_i : values.to_s
      Hash[keys.zip(values)]
    end
    puts JSON.pretty_generate(data)
  end

  def print_csv
    puts export_csv
  end

end

