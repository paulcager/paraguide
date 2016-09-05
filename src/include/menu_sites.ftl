<#list sites as site>
<li><a href="javascript: sites['${site.id}'].info()">${site.name} (${site.club})</a></li>
</#list>
