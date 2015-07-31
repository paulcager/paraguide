<#list sites as site>
<li><a href="javascript: sites['${site.id}'].info()">${site.name}</a></li>
</#list>
